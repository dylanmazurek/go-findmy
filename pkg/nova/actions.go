package nova

import (
	"context"
	"net/http"
	"time"

	"github.com/dylanmazurek/go-findmy/pkg/nova/constants"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (c *Client) ExecuteAction(ctx context.Context, canonicId string) error {
	requestUuid := uuid.New()

	log := log.Ctx(ctx).With().
		Str(constants.LOG_CLIENT_UUID, c.clientUuid).
		Str(constants.LOG_CANONIC_ID, canonicId).
		Str(constants.LOG_REQUEST_UUID, requestUuid.String()).
		Logger()

	log.Trace().Msg("executing action")

	lastHighTrafficEnablingTime := time.Now().Add(time.Duration(-5 * time.Hour)).Unix()

	var reqMessage = &bindings.ExecuteActionRequest{
		Action: &bindings.ExecuteActionType{
			LocateTracker: &bindings.ExecuteActionLocateTrackerType{
				LastHighTrafficEnablingTime: &bindings.Time{
					Seconds: uint32(lastHighTrafficEnablingTime),
				},
				ContributorType: bindings.SpotContributorType_FMDN_ALL_LOCATIONS,
			},
		},
		Scope: &bindings.ExecuteActionScope{
			Type: bindings.DeviceType_SPOT_DEVICE,
			Device: &bindings.ExecuteActionDeviceIdentifier{
				CanonicId: &bindings.CanonicId{
					Id: canonicId,
				},
			},
		},
		RequestMetadata: &bindings.ExecuteActionRequestMetadata{
			Type:          bindings.DeviceType_SPOT_DEVICE,
			RequestUuid:   requestUuid.String(),
			FmdClientUuid: c.clientUuid,
			Unknown:       true,
			GcmRegistrationId: &bindings.GcmCloudMessagingIdProtobuf{
				Id: *c.notifierSession.FcmSession.RegistrationToken,
			},
		},
	}

	req, err := c.NewRequest(ctx, http.MethodPost, constants.PATH_EXECUTE_ACTION, reqMessage, nil)
	if err != nil {
		return err
	}

	err = c.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
