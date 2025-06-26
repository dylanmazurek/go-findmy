package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dylanmazurek/go-findmy/pkg/notifier/constants"
	"github.com/dylanmazurek/go-findmy/pkg/notifier/models"
	shared "github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog/log"
)

type Session struct {
	Username      string             `json:"username"`
	AndroidId     *uint64            `json:"androidId"`
	SecurityToken *uint64            `json:"securityToken"`
	OwnerKey      *string            `json:"ownerKey"`
	FcmSession    *models.FcmSession `json:"fcmSession"`
	AdmSession    *models.AdmSession `json:"admSession"`
}

func (s *Session) GetEmail() string {
	email := fmt.Sprintf("%s@%s", s.Username, constants.GMAIL_DOMAIN)

	return email
}

func NewSession(ctx context.Context, sessionStr *string) (*Session, error) {
	log := log.Ctx(ctx).With().Str("client", constants.CLIENT_NAME).Logger()

	var session Session
	if sessionStr != nil {
		log.Info().Msg("session file set, loading session")

		err := session.LoadSession(ctx, *sessionStr)
		if err != nil {
			return nil, err
		}
	}

	if session.FcmSession == nil {
		session.FcmSession = &models.FcmSession{}
	}

	if session.AdmSession == nil {
		session.AdmSession = &models.AdmSession{}
	}

	if sessionStr == nil {
		log.Info().Msg("session file not set, creating new session")

		err := session.SaveSession(ctx, shared.DEFAULT_SESSION_FILE)
		if err != nil {
			return nil, err
		}
	}

	return &session, nil
}

func (s *Session) LoadSession(ctx context.Context, sessionStr string) error {
	log := log.Ctx(ctx)

	log.Debug().Msg("loading session")

	var session Session
	_, err := marshmallow.Unmarshal([]byte(sessionStr), &session)
	if err != nil {
		return err
	}

	s.Username = session.Username
	s.AndroidId = session.AndroidId
	s.SecurityToken = session.SecurityToken
	s.OwnerKey = session.OwnerKey
	s.FcmSession = session.FcmSession
	s.AdmSession = session.AdmSession

	log.Debug().Msg("session loaded")

	return nil
}

func (s *Session) SaveSession(ctx context.Context, f string) error {
	log := log.Ctx(ctx)

	log.Debug().Msg("saving session")

	jsonDetails, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	jsonDetailsBytes := []byte(jsonDetails)
	err = os.WriteFile(f, jsonDetailsBytes, 0644)
	if err != nil {
		return err
	}

	log.Debug().Msg("session saved")

	return nil
}
