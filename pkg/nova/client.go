package nova

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/google-findmy/pkg/shared/session"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	internalClient *http.Client

	clientUuid string
	session    *session.Session

	loggerCtx context.Context
}

func New(ctx context.Context, opts ...Option) (*Client, error) {
	clientOptions := DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	loger := zerolog.Ctx(ctx).With().Str("component", "nova-client").Logger()
	log.Logger = loger

	clientUuid := uuid.New()

	newClient := &Client{
		internalClient: http.DefaultClient,

		clientUuid: clientUuid.String(),
		session:    &clientOptions.session,

		loggerCtx: ctx,
	}

	if newClient.session.AdmSession.AdmToken == "" {
		_, err := newClient.getAdmToken()
		if err != nil {
			return nil, err
		}
	}

	authClient, err := createAuthTransport(newClient.session.AdmSession.AdmToken)
	if err != nil {
		return nil, err
	}

	newClient.internalClient = authClient

	return newClient, nil
}

func (c *Client) NewRequest(method string, path string, message proto.Message, params *url.Values) (*http.Request, error) {
	urlString := fmt.Sprintf("%s/%s", constants.NOVA_BASE_URL, path)
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

	return req, nil
}

func (c *Client) Do(req *http.Request, resp interface{}) error {
	httpResponse, err := c.internalClient.Do(req)
	if err != nil || httpResponse == nil || httpResponse.StatusCode >= 400 {
		if httpResponse != nil {
			log.Error().Err(err).
				Str("status", httpResponse.Status).
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
		return fmt.Errorf("response is not a proto.Message")
	}

	respBody := proto.Unmarshal(bodyBytes, protoMessage)
	if respBody != nil {
		return respBody
	}

	return nil
}
