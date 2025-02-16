package nova

import (
	"net/http"

	"github.com/dylanmazurek/google-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/google/uuid"
)

func (c *Client) GetDevices() (*bindings.DevicesList, error) {
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
	err = c.Do(req, &deviceList)
	if err != nil {
		return nil, err
	}

	return &deviceList, nil
}
