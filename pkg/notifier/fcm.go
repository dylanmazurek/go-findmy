package notifier

import (
	"context"
	"time"

	"github.com/dylanmazurek/go-findmy/pkg/notifier/constants"
	"github.com/dylanmazurek/go-findmy/pkg/notifier/models"
	shared "github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog/log"
)

type ReconnectionConfig struct {
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

func DefaultReconnectionConfig() *ReconnectionConfig {
	return &ReconnectionConfig{
		MaxRetries:        -1,
		InitialDelay:      5 * time.Second,
		MaxDelay:          5 * time.Minute,
		BackoffMultiplier: 2.0,
	}
}

func (s *Session) StartFCMWithReconnection(ctx context.Context) error {
	if s.reconnectConfig == nil {
		s.reconnectConfig = DefaultReconnectionConfig()
	}

	if s.stopReconnect == nil {
		s.stopReconnect = make(chan bool, 1)
	}

	if s.connectionMonitor == nil {
		s.connectionMonitor = make(chan error, 1)
	}

	go s.monitorConnection(ctx)

	return s.startFCMConnection(ctx, false)
}

func (s *Session) startFCMConnection(ctx context.Context, isReconnect bool) error {
	log := log.Ctx(ctx).With().Str("client", constants.CLIENT_NAME).Logger()

	client, err := s.NewFCMClient(ctx, false)
	if err != nil {
		if isReconnect {
			log.Error().Err(err).Msg("failed to create fcm client during reconnection")
			s.connectionMonitor <- err

			return nil
		}

		return err
	}

	s.fcmClient = client

	go func() {
		err := client.StartListening()
		if err != nil {
			log.Warn().Err(err).Msg("fcm start listening failed")
			s.connectionMonitor <- err
		}
	}()

	if isReconnect {
		log.Info().Msg("fcm reconnection successful")

		s.reconnectMutex.Lock()
		s.isReconnecting = false
		s.reconnectMutex.Unlock()

		return nil
	}

	log.Info().Msg("fcm connection established")

	return nil
}

func (s *Session) monitorConnection(ctx context.Context) {
	log := log.Ctx(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("context cancelled, stopping connection monitor")
			return
		case <-s.stopReconnect:
			log.Info().Msg("connection monitor stopped")
			return
		case err := <-s.connectionMonitor:
			log.Warn().Err(err).
				Msg("fcm connection error detected, scheduling reconnection")

			s.scheduleReconnection(ctx)
		}
	}
}

func (s *Session) scheduleReconnection(ctx context.Context) {
	s.reconnectMutex.Lock()

	if s.isReconnecting {
		s.reconnectMutex.Unlock()
		return
	}

	s.isReconnecting = true
	s.reconnectMutex.Unlock()

	go s.handleReconnection(ctx)
}

func (s *Session) handleReconnection(ctx context.Context) {
	log := log.Ctx(ctx)

	attempt := 0
	delay := s.reconnectConfig.InitialDelay

	for {
		select {
		case <-s.stopReconnect:
			log.Info().Msg("reconnection stopped")
			return
		case <-ctx.Done():
			log.Info().Msg("context cancelled, stopping reconnection")
			return
		default:
		}

		if s.reconnectConfig.MaxRetries > 0 && attempt >= s.reconnectConfig.MaxRetries {
			log.Error().Int("max_retries", s.reconnectConfig.MaxRetries).Msg("max reconnection attempts reached")
			return
		}

		attempt++
		log.Info().
			Int("attempt", attempt).
			Dur("delay", delay).
			Msg("attempting fcm reconnection")

		time.Sleep(delay)

		if s.fcmClient != nil {
			s.fcmClient.Close()
		}

		err := s.startFCMConnection(ctx, true)
		if err == nil {
			log.Info().Int("attempts", attempt).Msg("fcm reconnection successful")
			return
		}

		log.Error().Err(err).Int("attempt", attempt).Msg("fcm reconnection failed")

		delay = min(time.Duration(float64(delay)*s.reconnectConfig.BackoffMultiplier), s.reconnectConfig.MaxDelay)
	}
}

func (s *Session) StopFCMReconnection() {
	select {
	case s.stopReconnect <- true:
	default:
	}

	if s.fcmClient != nil {
		s.fcmClient.Close()
	}
}

func (s *Session) IsFCMConnected() bool {
	s.reconnectMutex.Lock()
	defer s.reconnectMutex.Unlock()

	return s.fcmClient != nil && !s.isReconnecting
}

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
	}

	return newSession, &c.AndroidId, &c.SecurityToken, err
}

func (s *Session) NewFCMClient(ctx context.Context, forceRegister bool) (*fcmreceiver.FCMClient, error) {
	log := log.Ctx(ctx).With().Str("client", "fcm").Logger()

	log.Trace().Msg("creating new fcm client")

	newClient := &fcmreceiver.FCMClient{
		ProjectID: constants.PROJECT_ID,
		AppId:     constants.APP_ID,
		ApiKey:    constants.API_KEY,
		AndroidApp: &fcmreceiver.AndroidFCM{
			GcmSenderId:    constants.MESSAGE_SENDER_ID,
			AndroidPackage: shared.ADM_APP_ID,
		},
	}

	err := s.prepareKeys(ctx, newClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare keys")

		return nil, err
	}

	if !forceRegister {
		log.Debug().Msg("using existing registration")

		newClient.FcmToken = *s.FcmSession.RegistrationToken
		newClient.AndroidId = *s.AndroidId
		newClient.SecurityToken = *s.SecurityToken

		return newClient, nil
	}

	log.Debug().Msg("refreshing registration")

	err = s.registerFCM(newClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to register fcm client")
		return nil, err
	}

	log.Trace().Msg("registered fcm client")

	return newClient, nil
}

func (s *Session) prepareKeys(ctx context.Context, c *fcmreceiver.FCMClient) error {
	log := log.Ctx(ctx)

	if s.FcmSession == nil || s.FcmSession.PrivateKeyBase64 == nil || s.FcmSession.AuthSecret == nil {
		s.FcmSession = &models.FcmSession{}

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

	log.Trace().Msg("loaded keys")

	return nil
}

func (s *Session) registerFCM(c *fcmreceiver.FCMClient) error {
	fcmToken, _, androidId, securityToken, err := c.Register()
	if err != nil {
		log.Error().Err(err).Msg("failed to register FCM client")
		return err
	}

	s.FcmSession.RegistrationToken = &fcmToken
	s.AndroidId = &androidId
	s.SecurityToken = &securityToken

	return nil
}
