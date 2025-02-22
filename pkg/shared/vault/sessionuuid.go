package vault

import (
	"fmt"
	"strings"

	"github.com/dylanmazurek/go-findmy/pkg/shared/uuid"
	"github.com/rs/zerolog/log"
)

func SessionSecretId(id string) (*string, error) {
	idHash := strings.ToLower(id)
	uuid, err := uuid.Parse(idHash)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse uuid")
		return nil, err
	}

	userHash, err := uuid.String()
	if err != nil {
		log.Error().Err(err).Msg("unable to convert uuid to string")
		return nil, err
	}

	sessionSecretId := fmt.Sprintf("session-%s", *userHash)

	return &sessionSecretId, nil
}
