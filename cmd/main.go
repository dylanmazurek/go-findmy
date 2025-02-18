package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dylanmazurek/google-findmy/internal/logger"
	"github.com/dylanmazurek/google-findmy/pkg/notifier"
	"github.com/dylanmazurek/google-findmy/pkg/nova"
	"github.com/dylanmazurek/google-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/google-findmy/pkg/shared/session"
	"github.com/markkurossi/tabulate"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	ctx = logger.InitLogger(ctx)

	sessionFile := constants.DEFAULT_SESSION_FILE
	session, err := session.New(ctx, &sessionFile)
	if err != nil {
		panic(err)
	}

	novaClient, notifyClient, err := initClients(ctx, session)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize clients")
	}

	go func() {
		log.Info().Msg("starting to listen for notifications")
		err = notifyClient.StartListening()
		if err != nil {
			log.Error().Err(err).Msg("failed to start listening")
		}
	}()

	err = printDevices(novaClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to print devices")
	}

	devices, err := novaClient.GetDevices()

	for _, device := range devices.DeviceMetadata {
		var canonicId string
		switch device.GetIdentifierInformation().GetType() {
		case bindings.IdentifierInformationType_IDENTIFIER_ANDROID:
			canonicId = device.GetIdentifierInformation().GetPhoneInformation().GetCanonicIds().GetCanonicId()[0].GetId()
		default:
			canonicId = device.GetIdentifierInformation().GetCanonicIds().GetCanonicId()[0].GetId()
		}

		log.Info().Str("canonicId", canonicId).Msg("executing action")

		err = novaClient.ExecuteAction(canonicId)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to execute action")
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("waiting for signals")

	<-sigs

	log.Info().Msg("received signal, stopping listener")
}

func initClients(ctx context.Context, session *session.Session) (*nova.Client, *notifier.Notifier, error) {
	log := log.Ctx(ctx)

	log.Info().Msg("initializing clients")

	notifyClient := notifier.NewNotifier(ctx)

	*session.FcmSession.RegistrationToken = notifyClient.GetFcmToken()

	clientOps := []nova.Option{
		nova.WithSession(*session),
	}

	novaClient, err := nova.New(ctx, clientOps...)
	if err != nil {
		return nil, nil, err
	}

	log.Info().Msg("clients initialized")

	return novaClient, notifyClient, nil
}

func printDevices(novaClient *nova.Client) error {
	devices, err := novaClient.GetDevices()
	if err != nil {
		return err
	}

	tab := tabulate.New(tabulate.ASCII)
	tab.Header("UUID")
	tab.Header("Name")

	for _, device := range devices.DeviceMetadata {
		var canonicId string
		switch device.GetIdentifierInformation().GetType() {
		case bindings.IdentifierInformationType_IDENTIFIER_ANDROID:
			canonicId = device.GetIdentifierInformation().GetPhoneInformation().GetCanonicIds().GetCanonicId()[0].GetId()
		default:
			canonicId = device.GetIdentifierInformation().GetCanonicIds().GetCanonicId()[0].GetId()
		}

		newRow := tab.Row()
		newRow.Column(canonicId)
		newRow.Column(device.GetUserDefinedDeviceName())
	}

	fmt.Println(tab.String())

	return nil
}
