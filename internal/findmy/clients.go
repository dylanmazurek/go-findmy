package findmy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	shared "github.com/dylanmazurek/go-findmy/pkg/shared/models"
	"github.com/dylanmazurek/go-findmy/pkg/shared/vault"
	"github.com/rs/zerolog/log"
)

func (s *Service) initClients(ctx context.Context) error {
	log := log.Ctx(ctx)

	log.Trace().Msg("initializing clients")

	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultAppRoleId := os.Getenv("VAULT_APPROLE_ID")
	vaultSecretId := os.Getenv("VAULT_SECRET_ID")

	vaultClient, err := vault.NewClient(ctx, vaultAddr, vaultAppRoleId, vaultSecretId)
	if err != nil {
		return err
	}

	vaultSecret, err := vaultClient.GetSecret(ctx, "kv", "go-findmy")
	if err != nil {
		return err
	}

	mqttUrl, hasMqttUrl := os.LookupEnv("MQTT_URL")
	if !hasMqttUrl {
		return fmt.Errorf("MQTT_URL not found in environment")
	}

	mqttUsername, ok := vaultSecret["MQTT_USERNAME"].(string)
	if !ok {
		return fmt.Errorf("MQTT_USERNAME not found in vault secret")
	}

	mqttPassword, ok := vaultSecret["MQTT_PASSWORD"].(string)
	if !ok {
		return fmt.Errorf("MQTT_PASSWORD not found in vault secret")
	}

	publisher, err := publisher.NewPublisher(ctx, mqttUrl, mqttUsername, mqttPassword)
	if err != nil {
		return err
	}

	sessionIrf, ok := vaultSecret["SESSION"].(map[string]any)
	if !ok {
		return fmt.Errorf("SESSION not found in vault secret")
	}

	sessionBytes, err := json.Marshal(sessionIrf)
	if err != nil {
		return err
	}

	var session *notifier.Session
	err = json.Unmarshal([]byte(sessionBytes), &session)
	if err != nil {
		return err
	}

	semanticLocationsIrf, ok := vaultSecret["SEMANTIC_LOCATIONS"].([]any)
	if !ok {
		return fmt.Errorf("SEMANTIC_LOCATIONS not found in vault secret")
	}

	semanticLocationsBytes, err := json.Marshal(semanticLocationsIrf)
	if err != nil {
		return err
	}

	var semanticLocations []shared.SemanticLocation
	err = json.Unmarshal([]byte(semanticLocationsBytes), &semanticLocations)
	if err != nil {
		return err
	}

	notifierClient, err := notifier.NewClient(ctx, *session, publisher, semanticLocations)
	if err != nil {
		return err
	}

	session.FcmSession.RegistrationToken = notifierClient.GetFcmToken()

	clientOps := []nova.Option{
		nova.WithNotifierSession(session),
	}

	novaClient, err := nova.NewClient(ctx, clientOps...)
	if err != nil {
		return err
	}

	log.Trace().Msg("clients initialized")

	s.novaClient = novaClient
	s.notifierClient = notifierClient
	s.publisherClient = publisher

	return nil
}
