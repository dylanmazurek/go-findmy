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
	"github.com/rs/zerolog/log"
)

func (s *Service) initClients(ctx context.Context, secrets map[string]any) error {
	log := log.Ctx(ctx)

	log.Trace().Msg("initializing clients")

	err := checkSecrets(secrets)
	if err != nil {
		return err
	}

	mqttUrl, hasMqttUrl := os.LookupEnv("MQTT_URL")
	if !hasMqttUrl {
		return fmt.Errorf("MQTT_URL not found in environment")
	}

	publisher, err := publisher.NewPublisher(ctx, mqttUrl, secrets["MQTT_USERNAME"].(string), secrets["MQTT_PASSWORD"].(string))
	if err != nil {
		return err
	}

	sessionBytes, err := json.Marshal(secrets["SESSION"].(map[string]any))
	if err != nil {
		return err
	}

	var session *notifier.Session
	err = json.Unmarshal([]byte(sessionBytes), &session)
	if err != nil {
		return err
	}

	semanticLocationsBytes, err := json.Marshal(secrets["SEMANTIC_LOCATIONS"].([]any))
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

func checkSecrets(secrets map[string]any) error {
	if secrets == nil {
		return fmt.Errorf("secrets cannot be nil")
	}

	requiredKeys := []string{"MQTT_USERNAME", "MQTT_PASSWORD", "SESSION", "SEMANTIC_LOCATIONS"}
	for _, key := range requiredKeys {
		if _, exists := secrets[key]; !exists {
			return fmt.Errorf("required secret %s not found", key)
		}
	}

	return nil
}
