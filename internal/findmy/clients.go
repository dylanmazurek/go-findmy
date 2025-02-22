package findmy

import (
	"context"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	"github.com/dylanmazurek/go-findmy/pkg/shared/session"
	"github.com/rs/zerolog/log"
)

func (f *FindMy) initClients(ctx context.Context, session *session.Session) error {
	log := log.Ctx(ctx)

	log.Trace().Msg("initializing clients")
	mqttUrl := os.Getenv("MQTT_URL")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	publisher, err := publisher.NewPublisher(ctx, mqttUrl, mqttUsername, mqttPassword)
	if err != nil {
		panic(err)
	}

	notifyClient := notifier.NewClient(ctx, session, publisher)
	*session.FcmSession.RegistrationToken = notifyClient.GetFcmToken()

	clientOps := []nova.Option{
		nova.WithSession(*session),
	}

	novaClient, err := nova.New(ctx, clientOps...)
	if err != nil {
		return err
	}

	log.Trace().Msg("clients initialized")

	f.novaClient = novaClient
	f.notifyClient = notifyClient
	f.publisherClient = publisher

	return nil
}
