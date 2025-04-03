package nova

import (
	"context"
	"net/http"

	"github.com/dylanmazurek/go-findmy/internal"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (c *Client) GetDevices(ctx context.Context) (*bindings.DevicesList, error) {
	log := log.Ctx(ctx)

	log.Info().Msg("getting devices")

	requestUuid := uuid.New()

	var reqMessage = &bindings.DevicesListRequest{
		DeviceListRequestPayload: &bindings.DevicesListRequestPayload{
			Id:   requestUuid.String(),
			Type: bindings.DeviceType_SPOT_DEVICE,
		},
	}

	req, err := c.NewRequest(http.MethodPost, constants.PATH_LIST_DEVICES, reqMessage, nil)
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

func (c *Client) RefreshAllDevices(ctx context.Context) error {
	log := log.Ctx(ctx)

	devices, err := c.GetDevices(ctx)
	if err != nil {
		return err
	}

	for _, device := range devices.DeviceMetadata {
		firstCanonicId, err := internal.FormatUniqueId(device)
		if err != nil {
			log.Error().Err(err).Msg("failed to format unique ID")
			continue
		}

		log.Trace().Str("canonicId", *firstCanonicId).Msg("executing action")

		err = c.ExecuteAction(ctx, *firstCanonicId)
		if err != nil {
			return err
		}
	}

	return nil
}
