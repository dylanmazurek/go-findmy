package notifier

import (
	"context"
	"encoding/base64"
	"os"
	"slices"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/dylanmazurek/go-findmy/pkg/decryptor"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	shared "github.com/dylanmazurek/go-findmy/pkg/shared/models"
	"github.com/dylanmazurek/go-findmy/pkg/shared/session"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

var semanticLocations []shared.SemanticLocation

type Client struct {
	internalClient *fcmreceiver.FCMClient
	decryptor      *decryptor.Decryptor
	publisher      *publisher.Client

	session     *session.Session
	publishMqtt bool
}

func NewClient(ctx context.Context, s *session.Session, p *publisher.Client, sl []shared.SemanticLocation) (*Client, error) {
	log := log.Ctx(ctx)

	log.Trace().Msg("creating new notifier")

	fcmClient, err := s.NewFCMClient(ctx, true)
	if err != nil {
		log.Error().Err(err).Msg("failed to create FCM client")
		return nil, err
	}

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

	newNotifier := Client{
		internalClient: fcmClient,
		decryptor:      newDecryptor,
		publisher:      p,

		session:     s,
		publishMqtt: publishMqttEnv == "true",
	}

	return &newNotifier, nil
}

func (n *Client) StartListening(ctx context.Context) error {
	n.internalClient.OnDataMessage = func(message []byte) { n.onDataMessage(ctx, message) }
	n.internalClient.OnRawMessage = func(message *fcmreceiver.DataMessageStanza) { n.OnRawMessage(ctx, message) }

	go func(ctx context.Context) {
		log := log.Ctx(ctx)

		log.Info().Msg("starting listener")
		err := n.internalClient.StartListening()
		if err != nil {
			log.Error().Err(err).Msg("failed to start listening")
		}
	}(ctx)

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
		log.Error().Msg("failed to find FCM payload")
		return
	}

	fcmPayloadHex := appData[fcmPayloadIdx].GetValue()

	fcmPayload, err := base64.StdEncoding.DecodeString(fcmPayloadHex)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode FCM payload")
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
		log.Error().Msg("no locations found")
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

func (n *Client) GetFcmToken() string {
	fcmToken := n.internalClient.FcmToken
	return fcmToken
}

func (n *Client) StopListening() {
	n.internalClient.Close()
}
