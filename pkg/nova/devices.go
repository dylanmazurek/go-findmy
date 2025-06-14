package nova

import (
	"context"
	"net/http"

	"github.com/dylanmazurek/go-findmy/internal"
	"github.com/dylanmazurek/go-findmy/pkg/nova/constants"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (c *Client) GetDevices(ctx context.Context) (*bindings.DevicesList, error) {
	log := log.Ctx(ctx)

	log.Debug().Msg("fetching devices")

	requestUuid := uuid.New()

	var reqMessage = &bindings.DevicesListRequest{
		DeviceListRequestPayload: &bindings.DevicesListRequestPayload{
			Id:   requestUuid.String(),
			Type: bindings.DeviceType_SPOT_DEVICE,
		},
	}

	req, err := c.NewRequest(ctx, http.MethodPost, constants.PATH_LIST_DEVICES, reqMessage, nil)
	if err != nil {
		return nil, err
	}

	var deviceList bindings.DevicesList
	err = c.Do(ctx, req, &deviceList)
	if err != nil {
		return nil, err
	}

	return &deviceList, nil
}

func (c *Client) RefreshDevices(ctx context.Context) error {
	log := log.Ctx(ctx)

	log.Debug().Msg("refreshing devices")

	devices, err := c.GetDevices(ctx)
	if err != nil {
		return err
	}

	for _, device := range devices.DeviceMetadata {
		firstCanonicId, err := internal.FormatUniqueId(device)
		if err != nil {
			continue
		}

		log.Trace().
			Str(constants.LOG_CANONIC_ID, *firstCanonicId).
			Msg("executing action")

		err = c.ExecuteAction(ctx, *firstCanonicId)
		if err != nil {
			return err
		}
	}

	return nil
}
