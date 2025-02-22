package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/go-findmy/pkg/shared/session/models"
	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog"
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
	email := fmt.Sprintf("%s@gmail.com", s.Username)
	return email
}

// New creates a new session
// ctx is the context, f is the file to save the session to
func New(ctx context.Context, f *string) (*Session, error) {
	loger := zerolog.Ctx(ctx).With().Str("component", "session").Logger()
	log.Logger = loger

	var session Session
	if f != nil {
		log.Info().Msg("session file set, loading session")

		err := session.LoadSession(*f)
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

	if f == nil {
		log.Info().Msg("session file not set, creating new session")

		err := session.SaveSession(ctx, constants.DEFAULT_SESSION_FILE)
		if err != nil {
			return nil, err
		}
	}

	return &session, nil
}

// loadSession loads a session string
func (s *Session) LoadSession(sessionStr string) error {
	var session Session
	_, err := marshmallow.Unmarshal([]byte(sessionStr), &session)
	if err != nil {
		return err
	}

	*s = session

	return nil
}

// saveSession encodes a session and saves it to a file
func (s *Session) SaveSession(ctx context.Context, f string) error {
	log := log.Ctx(ctx)

	jsonDetails, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	jsonDetailsBytes := []byte(jsonDetails)
	err = os.WriteFile(f, jsonDetailsBytes, 0644)
	if err != nil {
		return err
	}

	log.Trace().Msg("session saved")

	return nil
}
