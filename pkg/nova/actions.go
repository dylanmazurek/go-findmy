package nova

import (
	"context"
	"net/http"
	"time"

	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/google/uuid"
)

func (c *Client) ExecuteAction(ctx context.Context, canonicId string) error {
	requestUuid := uuid.New()

	lastHighTrafficEnablingTime := time.Now().Add(time.Duration(-5 * time.Hour)).Unix()

	var reqMessage = &bindings.ExecuteActionRequest{
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
				Id: *c.session.FcmSession.RegistrationToken,
			},
		},
		Action: &bindings.ExecuteActionType{
			LocateTracker: &bindings.ExecuteActionLocateTrackerType{
				LastHighTrafficEnablingTime: &bindings.Time{
					Seconds: uint32(lastHighTrafficEnablingTime),
				},
				ContributorType: bindings.SpotContributorType_FMDN_ALL_LOCATIONS,
			},
		},
	}

	req, err := c.NewRequest(http.MethodPost, constants.PATH_EXECUTE_ACTION, reqMessage, nil)
	if err != nil {
		return err
	}

	err = c.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
