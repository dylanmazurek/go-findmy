package spot

import (
	"fmt"
	"net/http"

	"github.com/dylanmazurek/go-findmy/pkg/nova/models"
)

type addAuthHeaderTransport struct {
	T          http.RoundTripper
	AuthGetter func() *models.Auth
}

func (adt *addAuthHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	auth := adt.AuthGetter()

	bearerAuth := fmt.Sprintf("Bearer %s", auth.Token)
	req.Header.Add("Authorization", bearerAuth)
	// req.Header.Add("Accept-Language", constants.NOVA_API_LANGUAGE)
	// req.Header.Add("User-Agent", constants.NOVA_USER_AGENT)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	return adt.T.RoundTrip(req)
}

func (c *Client) createAuthTransport() (*http.Client, error) {
	authClient := &http.Client{
		Transport: &addAuthHeaderTransport{
			T: http.DefaultTransport,
		},
	}

	return authClient, nil
}
