package session

import (
	"context"

	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/google-findmy/pkg/shared/session/models"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog/log"
)

func NewSessionFromClient(c *fcmreceiver.FCMClient) (*models.FcmSession, *uint64, *uint64, error) {
	privateKey, err := c.GetPrivateKeyBase64()
	if err != nil {
		return nil, nil, nil, err
	}

	authSecret := c.GetAuthSecretBase64()

	newSession := &models.FcmSession{
		RegistrationToken: &c.FcmToken,
		PrivateKeyBase64:  &privateKey,
		AuthSecret:        &authSecret,
		//PersistentIds:         c.PersistentIds,
		//InstallationAuthToken: c.InstallationAuthToken,
	}

	return newSession, &c.AndroidId, &c.SecurityToken, err
}

func (s *Session) NewFCMClient(ctx context.Context, refreshReg bool) (*fcmreceiver.FCMClient, error) {
	log := log.Ctx(ctx)

	log.Info().Msg("creating new fcm client")

	newClient := fcmreceiver.FCMClient{
		ProjectID: constants.PROJECT_ID,
		AppId:     constants.APP_ID,
		ApiKey:    constants.API_KEY,
		AndroidApp: &fcmreceiver.AndroidFCM{
			GcmSenderId:    constants.MESSAGE_SENDER_ID,
			AndroidPackage: constants.APP_ADM,
		},
	}

	err := s.prepareKeys(ctx, &newClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare keys")
		return nil, err
	}

	if !refreshReg {
		log.Info().Msg("using existing registration")

		newClient.FcmToken = *s.FcmSession.RegistrationToken
		newClient.AndroidId = *s.AndroidId
		newClient.SecurityToken = *s.SecurityToken
	}

	if refreshReg {
		log.Info().Msg("refreshing registration")

		err = s.registerFCM(&newClient)
		if err != nil {
			log.Error().Err(err).Msg("failed to register FCM client")
			return nil, err
		}
	}

	return &newClient, nil
}

func (s *Session) prepareKeys(ctx context.Context, c *fcmreceiver.FCMClient) error {
	log := log.Ctx(ctx)

	if s.FcmSession.PrivateKeyBase64 == nil || s.FcmSession.AuthSecret == nil {
		privateKey, authSecret, err := c.CreateNewKeys()
		if err != nil {
			return err
		}

		s.FcmSession.PrivateKeyBase64 = &privateKey
		s.FcmSession.AuthSecret = &authSecret
	}

	err := c.LoadKeys(*s.FcmSession.PrivateKeyBase64, *s.FcmSession.AuthSecret)
	if err != nil {
		return err
	}

	log.Info().Msg("loaded keys")

	return nil
}

func (s *Session) registerFCM(c *fcmreceiver.FCMClient) error {
	fcmToken, _, androidId, securityToken, err := c.Register()
	if err != nil {
		log.Error().Err(err).Msg("failed to register FCM client")
		return err
	}

	s.FcmSession.RegistrationToken = &fcmToken
	//session.FcmSession.GcmToken = gcmToken
	s.AndroidId = &androidId
	s.SecurityToken = &securityToken

	return nil
}
