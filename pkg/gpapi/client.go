package gpapi

import "net/http"

type Client struct {
	internalClient http.Client
}

func NewClient() *Client {
	newClient := &Client{
		internalClient: *http.DefaultClient,
	}

	return newClient
}
