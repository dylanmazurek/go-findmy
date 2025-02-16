package spot

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const BASE_URL = "https://partner-devices-pa.googleapis.com"

type Client struct {
	ctx            context.Context
	internalClient *http.Client
	token          string
}

func New(ctx context.Context, token string) (*Client, error) {
	httpClient := &http.Client{}

	return &Client{
		ctx:            ctx,
		internalClient: httpClient,
		token:          token,
	}, nil
}

func (c *Client) Do(method string, path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", BASE_URL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	httpResponse, err := c.internalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", httpResponse.StatusCode)
	}

	return io.ReadAll(httpResponse.Body)
}

func (c *Client) CreateBleDevice(deviceId string, ownershipToken string) error {
	path := fmt.Sprintf("spot/v1/createbledevice?device_id=%s&ownership_token=%s",
		deviceId, ownershipToken)

	_, err := c.Do("POST", path, nil)
	return err
}

func (c *Client) GetEidInfo(deviceId string) ([]byte, error) {
	path := fmt.Sprintf("spot/v1/gete2eeinfo?device_id=%s", deviceId)
	return c.Do("GET", path, nil)
}
