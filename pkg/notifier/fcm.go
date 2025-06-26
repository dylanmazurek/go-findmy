package notifier

import (
	"context"

	"github.com/dylanmazurek/go-findmy/pkg/notifier/models"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog/log"
)

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

	log.Debug().Msg("loaded keys")

	return nil
}
