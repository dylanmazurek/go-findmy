package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Client struct {
	internalClient *autopaho.ConnectionManager
}

func NewPublisher(ctx context.Context, mqttUrl string, username string, password string) (*Client, error) {
	clientId := "google-findmy"

	u, err := url.Parse(mqttUrl)
	if err != nil {
		return nil, err
	}

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60,
		ClientConfig: paho.ClientConfig{
			ClientID: clientId,
		},
	}

	cliCfg.ConnectUsername = username
	cliCfg.ConnectPassword = []byte(password)

	c, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		return nil, err
	}

	if err = c.AwaitConnection(ctx); err != nil {
		return nil, err
	}

	newClient := Client{
		internalClient: c,
	}

	return &newClient, nil
}

func (c *Client) AddDevice(ctx context.Context, device models.Device) (*paho.PublishResponse, error) {
	deviceJson, err := json.MarshalIndent(device, "", " ")
	if err != nil {
		return nil, err
	}

	payload := &paho.Publish{
		QoS:     1,
		Topic:   device.GetConfigTopic(),
		Retain:  true,
		Payload: deviceJson,
	}

	resp, err := c.internalClient.Publish(ctx, payload)
	if err != nil {
		return resp, err
	}

	return resp, err
}

func (c *Client) UpdateTracker(ctx context.Context, report models.Report) (*paho.PublishResponse, error) {
	deviceJson, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("findmy2mqtt/%s/attributes", report.UniqueId)

	payload := &paho.Publish{
		QoS:     1,
		Topic:   topic,
		Payload: deviceJson,
	}

	resp, err := c.internalClient.Publish(ctx, payload)
	if err != nil {
		return resp, err
	}

	return resp, err
}
