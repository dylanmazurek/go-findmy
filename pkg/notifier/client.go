package notifier

import (
	"context"
	"encoding/base64"
	"os"
	"slices"

	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/google-findmy/pkg/shared/session"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Notifier struct {
	internalClient *fcmreceiver.FCMClient

	loggerCtx context.Context
}

func NewNotifier(ctx context.Context) *Notifier {
	loger := zerolog.Ctx(ctx).With().Str("component", "notifier").Logger()
	log.Logger = loger

	log.Info().Msg("creating new notifier")

	sessionFile := constants.DEFAULT_SESSION_FILE
	session, err := session.New(ctx, &sessionFile)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		return nil
	}

	fcmClient, err := session.NewFCMClient(ctx, true)
	if err != nil {
		log.Error().Err(err).Msg("failed to create FCM client")
		return nil
	}

	err = session.SaveSession(ctx, constants.DEFAULT_SESSION_FILE)
	if err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return nil
	}

	newNotifier := Notifier{
		internalClient: fcmClient,

		loggerCtx: ctx,
	}

	return &newNotifier
}

func (n *Notifier) StartListening() error {
	log := log.Ctx(n.loggerCtx)

	n.internalClient.OnDataMessage = func(message []byte) {
		log.Info().Bytes("msg", message).Msg("received data message")
	}

	n.internalClient.OnRawMessage = func(message *fcmreceiver.DataMessageStanza) {
		appData := message.GetAppData()

		fcmPayloadIdx := slices.IndexFunc(appData, func(i *fcmreceiver.AppData) bool {
			key := i.GetKey()
			isFcmPayload := (key == "com.google.android.apps.adm.FCM_PAYLOAD")

			return isFcmPayload
		})

		if fcmPayloadIdx == -1 {
			log.Error().Msg("failed to find FCM payload")
			return
		}

		fcmPayloadHex := appData[fcmPayloadIdx].GetValue()

		fcmPayload, err := base64.RawStdEncoding.DecodeString(fcmPayloadHex)
		if err != nil {
			log.Error().Err(err).Msg("failed to decode FCM payload")
			return
		}

		err = os.WriteFile("message.txt", fcmPayload, 0644)
		if err != nil {
			log.Error().Err(err).Msg("failed to write message to file")
		}

		log.Info().
			Str("id", *message.Id).
			Str("from", message.GetFrom()).
			Str("category", *message.Category).
			Msg("received raw message")
	}

	go func() {
		err := n.internalClient.StartListening()
		if err != nil {
			log.Error().Err(err).Msg("failed to start listening")
		}
	}()

	log.Info().Msg("listening for messages")

	return nil
}

func (n *Notifier) GetFcmToken() string {
	fcmToken := n.internalClient.FcmToken
	return fcmToken
}
