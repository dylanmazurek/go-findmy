package notifier

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/dylanmazurek/go-findmy/pkg/decryptor"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/go-findmy/pkg/shared/session"
	fcmreceiver "github.com/morhaviv/go-fcm-receiver"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	internalClient *fcmreceiver.FCMClient
	decryptor      *decryptor.Decryptor
	publisher      *publisher.Client

	semanticLocations map[string]models.Attributes

	session *session.Session
	ctx     *context.Context
}

func NewClient(ctx context.Context, s *session.Session, p *publisher.Client) (*Client, error) {
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

	semanticLocations, err := getSemanticLocations()
	if err != nil {
		log.Error().Err(err).Msg("failed to get semantic locations")
		return nil, err
	}

	newNotifier := Client{
		internalClient: fcmClient,
		decryptor:      newDecryptor,
		publisher:      p,

		semanticLocations: semanticLocations,

		session: s,
		ctx:     &ctx,
	}

	return &newNotifier, nil
}

// StartListening starts listening for messages
// this is a non-blocking call
func (n *Client) StartListening(ctx context.Context) error {
	n.internalClient.OnDataMessage = n.onDataMessage
	n.internalClient.OnRawMessage = n.OnRawMessage

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

func (n *Client) onDataMessage(message []byte) {
	log := log.Ctx(*n.ctx)
	log.Trace().Bytes("msg", message).Msg("received data message")
}

var uniqueIdRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func (n *Client) OnRawMessage(message *fcmreceiver.DataMessageStanza) {
	log := log.Ctx(*n.ctx)

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

	ctx := context.Background()
	locations, err := n.decryptor.DecryptDeviceUpdate(ctx, &deviceUpdate)
	if err != nil {
		log.Error().Err(err).Msg("failed to decrypt device update")
		return
	}

	for _, loc := range locations {
		log.Info().Interface("location", loc).Msg("location")
		if deviceUpdate.GetDeviceMetadata().GetIdentifierInformation().GetType() != bindings.IdentifierInformationType_IDENTIFIER_ANDROID {
			deviceName := deviceUpdate.GetDeviceMetadata().GetUserDefinedDeviceName()

			canonicId := deviceUpdate.GetDeviceMetadata().GetIdentifierInformation().GetCanonicIds().GetCanonicId()[0].GetId()
			canonicIdSplit := strings.Split(canonicId, "-")
			serial := canonicIdSplit[len(canonicIdSplit)-1]
			uniqueId := uniqueIdRegex.ReplaceAllString(fmt.Sprintf("%s_%s", deviceName, serial), "_")
			uniqueId = strings.ToLower(uniqueId)
			isSemanticLocation := (loc.SemanticLocation != "")

			var attributes models.Attributes
			if isSemanticLocation {
				semanticLocation, ok := n.semanticLocations[strings.ToLower(loc.SemanticLocation)]
				if ok {
					attributes.Latitude = semanticLocation.Latitude
					attributes.Longitude = semanticLocation.Longitude
					attributes.Accuracy = 10
				} else {
					log.Warn().Str("semantic_location", loc.SemanticLocation).Msg("semantic location not found")
				}
			}

			if !isSemanticLocation {
				attributes.Latitude = float32(loc.Latitude)
				attributes.Longitude = float32(loc.Longitude)
				attributes.Altitude = float32(loc.Altitude)

				accuracy := deviceUpdate.GetDeviceMetadata().GetInformation().GetLocationInformation().GetReports().RecentLocationAndNetworkLocations.GetRecentLocation().GetGeoLocation().GetAccuracy()
				attributes.Accuracy = accuracy
			}

			n.publisher.UpdateTracker(ctx, uniqueId, attributes)
		}
	}
}

func (n *Client) GetFcmToken() string {
	fcmToken := n.internalClient.FcmToken
	return fcmToken
}

func (n *Client) StopListening() {
	n.internalClient.Close()
}

func getSemanticLocations() (map[string]models.Attributes, error) {
	semanticLocationsFile, err := os.ReadFile("semantic_locations.json")
	if err != nil {
		log.Error().Err(err).Msg("failed to read semantic locations")
		return nil, err
	}

	var semanticLocations struct {
		Locations []struct {
			Names     []string `json:"names"`
			Longitude float32  `json:"longitude"`
			Latitude  float32  `json:"latitude"`
		} `json:"locations"`
	}
	json.Unmarshal(semanticLocationsFile, &semanticLocations)

	var semanticLocationsMap = make(map[string]models.Attributes)
	for _, location := range semanticLocations.Locations {
		for _, name := range location.Names {
			semanticLocationsMap[strings.ToLower(name)] = models.Attributes{
				Longitude: location.Longitude,
				Latitude:  location.Latitude,
			}
		}
	}

	return semanticLocationsMap, nil
}
