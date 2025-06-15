package spot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dylanmazurek/go-findmy/pkg/spot/constants"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	internalClient *http.Client
}

func New(ctx context.Context, opts ...Option) (*Client, error) {
	log := log.Ctx(ctx).With().Str("client", "spot").Logger()

	log.Debug().Msg("creating new spot client")

	clientOptions := DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	newClient := &Client{
		internalClient: &http.Client{
			Transport: &http.Transport{
				ForceAttemptHTTP2: false,
			},
		},
	}

	return newClient, nil
}

func (c *Client) NewRequest(method string, path string, message proto.Message, params *url.Values) (*http.Request, error) {
	urlString := fmt.Sprintf("%s/%s", constants.API_BASE_URL, path)
	requestUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	if params != nil {
		requestUrl.RawQuery = params.Encode()
	}

	var body io.Reader
	if message != nil {
		reqBody, err := proto.Marshal(message)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest(method, requestUrl.String(), body)
	if err != nil {
		return nil, err
	}

	// TODO: fix this code
	// tokenValid := c.auth.IsValid()
	// if !tokenValid {
	// 	log.Info().
	// 		Msg("adm token invalid, refreshing")

	// 	err = c.refreshAdmToken()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, resp interface{}) error {
	log := log.Ctx(ctx)

	httpResponse, err := c.internalClient.Do(req)
	if err != nil || httpResponse == nil || httpResponse.StatusCode >= 400 {
		if httpResponse != nil {
			log.Error().Err(err).
				Str(constants.LOG_HTTP_STATUS, httpResponse.Status).
				Msg("failed to execute request")
		}

		return err
	}
	defer httpResponse.Body.Close()

	bodyBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	if resp == nil {
		return nil
	}

	protoMessage, ok := resp.(proto.Message)
	if !ok {
		return fmt.Errorf("response type %T does not implement proto.Message", resp)
	}

	respBody := proto.Unmarshal(bodyBytes, protoMessage)
	if respBody == nil {
		return nil
	}

	return respBody
}
