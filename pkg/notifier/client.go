package notifier

import (
	"context"
	"encoding/base64"
	"os"
	"slices"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/dylanmazurek/go-findmy/pkg/decryptor"
	"github.com/dylanmazurek/go-findmy/pkg/notifier/constants"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	shared "github.com/dylanmazurek/go-findmy/pkg/shared/models"

	sharedconstants "github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

var semanticLocations []shared.SemanticLocation

type Client struct {
	internalClient *fcmreceiver.FCMClient

	decryptor *decryptor.Decryptor
	publisher *publisher.Client

	session     *Session
	publishMqtt bool
}

func NewClient(ctx context.Context, s Session, p *publisher.Client, sl []shared.SemanticLocation) (*Client, error) {
	log := log.Ctx(ctx).With().Str("client", constants.CLIENT_NAME).Logger()

	log.Trace().Msg("creating")

	newDecryptor, err := decryptor.NewDecryptor(s.OwnerKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to create decryptor")

		return nil, err
	}

	publishMqttEnv, hasPublishMqtt := os.LookupEnv("PUBLISH_MQTT")
	if !hasPublishMqtt {
		publishMqttEnv = "false"
	}

	semanticLocations = sl

	log.Trace().Msg("initializing internal client")
	newInternalClient, err := newInternalClient(ctx, s, true)
	if err != nil {
		log.Error().Err(err).Msg("failed to create internal client")
		return nil, err
	}

	newNotifier := &Client{
		decryptor: newDecryptor,
		publisher: p,

		session:        &s,
		internalClient: newInternalClient,
		publishMqtt:    (publishMqttEnv == "true"),
	}

	return newNotifier, nil
}

func newInternalClient(ctx context.Context, session Session, forceRegister bool) (*fcmreceiver.FCMClient, error) {
	log := log.Ctx(ctx).With().Str("client", "fcm").Logger()

	log.Trace().Msg("creating new fcm client")

	newClient := &fcmreceiver.FCMClient{
		ProjectID: constants.PROJECT_ID,
		AppId:     constants.APP_ID,
		ApiKey:    constants.API_KEY,
		AndroidApp: &fcmreceiver.AndroidFCM{
			GcmSenderId:    constants.MESSAGE_SENDER_ID,
			AndroidPackage: sharedconstants.ADM_APP_ID,
		},
	}

	err := session.prepareKeys(ctx, newClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to prepare keys")

		return nil, err
	}

	if !forceRegister {
		log.Debug().Msg("using existing registration")

		newClient.FcmToken = *session.FcmSession.RegistrationToken
		newClient.AndroidId = *session.AndroidId
		newClient.SecurityToken = *session.SecurityToken

		return newClient, nil
	}

	log.Debug().Msg("refreshing registration")

	fcmToken, _, androidId, securityToken, err := newClient.Register()
	if err != nil {
		log.Error().Err(err).Msg("failed to register FCM client")
		return nil, err
	}

	session.FcmSession.RegistrationToken = &fcmToken
	session.AndroidId = &androidId
	session.SecurityToken = &securityToken

	log.Trace().Msg("registered fcm client")

	return newClient, nil
}

func (n *Client) StartListening(ctx context.Context) error {
	log := log.Ctx(ctx).With().Str("client", constants.CLIENT_NAME).Logger()

	n.internalClient.OnDataMessage = func(message []byte) {
		n.onDataMessage(ctx, message)
	}

	n.internalClient.OnRawMessage = func(message *fcmreceiver.DataMessageStanza) {
		n.OnRawMessage(ctx, message)
	}

	go func() {
		log.Debug().Msg("starting fcm client")

		for {
			err := n.internalClient.StartListening()
			if err != nil {
				log.Error().Err(err).Msg("fcm start listening failed")
			}

			log.Debug().Msg("restarting fcm client")

			// Recreate the client to reset the connection
			newClient, err := newInternalClient(ctx, *n.session, true)
			if err != nil {
				log.Error().Err(err).Msg("failed to recreate fcm client")
				return
			}

			n.internalClient = newClient
		}
	}()

	return nil
}

func (n *Client) onDataMessage(ctx context.Context, message []byte) {
	log := log.Ctx(ctx)

	log.Trace().Bytes("msg", message).Msg("received data message")
}

func (n *Client) OnRawMessage(ctx context.Context, message *fcmreceiver.DataMessageStanza) {
	log := log.Ctx(ctx)

	log.Trace().Msg("received raw message")

	appData := message.GetAppData()

	fcmPayloadIdx := slices.IndexFunc(appData, func(i *fcmreceiver.AppData) bool {
		key := i.GetKey()
		isFcmPayload := (key == constants.MESSAGE_FCM_PAYLOAD_NAME)

		return isFcmPayload
	})

	if fcmPayloadIdx == -1 {
		log.Error().Msg("failed to find fcm payload")

		return
	}

	fcmPayloadHex := appData[fcmPayloadIdx].GetValue()

	fcmPayload, err := base64.StdEncoding.DecodeString(fcmPayloadHex)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode fcm payload")
		return
	}

	var deviceUpdate bindings.DeviceUpdate
	err = proto.Unmarshal([]byte(fcmPayload), &deviceUpdate)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal device update")
		return
	}

	locations, err := n.decryptor.DecryptDeviceUpdate(ctx, &deviceUpdate)
	if err != nil {
		log.Error().Err(err).Msg("failed to decrypt device update")
		return
	}

	if len(locations) == 0 {
		log.Warn().
			Str(constants.LOG_USER_DEFINED_DEVICE_NAME, deviceUpdate.DeviceMetadata.UserDefinedDeviceName).
			Msg("no recent locations found for device")

		return
	}

	var latestReport *shared.LocationReport
	for _, loc := range locations {
		locationReport, err := n.handleReport(ctx, &deviceUpdate, loc)
		if err != nil {
			log.Error().Err(err).Msg("failed to handle message")
			continue
		}

		if latestReport == nil || locationReport.ReportTime.After(latestReport.ReportTime) {
			latestReport = locationReport
		}
	}

	if latestReport == nil {
		log.Error().Msg("latest report is nil")
		return
	}

	if n.publishMqtt {
		pubReport := models.Report{
			UniqueId:  *latestReport.UniqueId,
			Latitude:  latestReport.Latitude,
			Longitude: latestReport.Longitude,
			Altitude:  latestReport.Altitude,
			Accuracy:  latestReport.Accuracy,
		}

		_, err := n.publisher.UpdateTracker(ctx, pubReport)
		if err != nil {
			log.Error().Err(err).Msg("failed to publish update")
		}

		log.Debug().Str("unique_id", pubReport.UniqueId).Msg("published location update")
	}

	deviceName := deviceUpdate.DeviceMetadata.GetUserDefinedDeviceName()

	log.Info().
		Str("name", deviceName).
		Int("count", len(locations)).
		Float64("latitude", latestReport.Latitude).
		Float64("longitude", latestReport.Longitude).
		Float64("accuracy", latestReport.Accuracy).
		Msg("report")
}

func (n *Client) GetFcmToken() *string {
	if n.internalClient != nil {
		return &n.internalClient.FcmToken
	}

	return nil
}
