package nova

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova/constants"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	internalClient *http.Client

	clientUuid string
	auth       *models.Auth

	notifierSession *notifier.Session
}

func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	log := log.Ctx(ctx).With().Str("client", constants.CLIENT_NAME).Logger()

	clientOptions := DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	newClientUuid := uuid.New()

	log = log.With().
		Str(constants.LOG_CLIENT_UUID, newClientUuid.String()).
		Logger()

	newClient := &Client{
		internalClient: http.DefaultClient,

		clientUuid: newClientUuid.String(),

		notifierSession: clientOptions.notifierSession,
	}

	err := newClient.validateAdmToken(ctx)
	if err == ErrTokenExpired {
		log.Debug().
			Msg("adm token expired, refreshing")

		err = newClient.refreshAdmToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	authClient, err := newClient.createAuthTransport()
	if err != nil {
		return nil, err
	}

	newClient.internalClient = authClient

	return newClient, nil
}

func (c *Client) NewRequest(ctx context.Context, method string, path string, message proto.Message, params *url.Values) (*http.Request, error) {
	log := log.Ctx(ctx)

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

	if body == nil {
		body = bytes.NewBuffer([]byte{})
	}

	requestLog := log.Trace().
		Str(constants.LOG_HTTP_METHOD, method).
		Str(constants.LOG_HTTP_URL, requestUrl.String())

	if message != nil {
		requestLog = requestLog.
			Str(constants.LOG_MESSAGE_TYPE, fmt.Sprintf("%T", message))
	}

	requestLog.Msg("creating new request")

	req, err := http.NewRequest(method, requestUrl.String(), body)
	if err != nil {
		return nil, err
	}

	tokenValid := c.auth.IsValid()
	if !tokenValid {
		log.Info().
			Msg("android device manager token invalid")

		err = c.refreshAdmToken(ctx)
		if err != nil {
			return nil, err
		}

	}

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
		return ErrResponseNotProtoMessage
	}

	respBody := proto.Unmarshal(bodyBytes, protoMessage)
	if respBody == nil {
		return nil
	}

	return respBody
}
