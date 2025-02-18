package nova

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dylanmazurek/google-findmy/pkg/nova/models"
	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog/log"
)

type addAuthHeaderTransport struct {
	T     http.RoundTripper
	Token *string
}

func (adt *addAuthHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	bearerAuth := fmt.Sprintf("Bearer %s", *adt.Token)
	req.Header.Add("Authorization", bearerAuth)
	req.Header.Add("Accept-Language", constants.NOVA_API_LANGUAGE)
	req.Header.Add("User-Agent", constants.NOVA_USER_AGENT)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	return adt.T.RoundTrip(req)
}

func createAuthTransport(token string) (*http.Client, error) {
	authClient := &http.Client{
		Transport: &addAuthHeaderTransport{
			T:     http.DefaultTransport,
			Token: &token,
		},
	}

	return authClient, nil
}

func (c *Client) validateAdmToken() error {
	urlPath := fmt.Sprintf("%s?access_token=%s", constants.GOOGLE_TOKEN_INFO_URL, c.session.AdmSession.AdmToken)
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenInfo models.TokenInfo
	_, err = marshmallow.Unmarshal(bodyBytes, &tokenInfo)
	if err != nil {
		return err
	}

	if tokenInfo.ExpiresIn < 60 {
		return ErrTokenExpired
	}

	return nil
}

func (c *Client) getAdmToken() (*string, error) {
	var service = fmt.Sprintf("oauth2:https://www.googleapis.com/auth/%s", constants.NOVA_CLIENT_SCOPE)

	formData := url.Values{}
	formData.Set("accountType", "HOSTED_OR_GOOGLE")
	formData.Set("Email", c.session.GetEmail())
	formData.Set("has_permission", "1")
	formData.Set("EncryptedPasswd", c.session.AdmSession.AasToken)
	formData.Set("service", service)
	formData.Set("source", constants.NOVA_CLIENT_SOURCE)
	formData.Set("androidId", fmt.Sprintf("%d", c.session.AndroidId))
	formData.Set("app", constants.APP_ADM)
	formData.Set("client_sig", constants.NOVA_CLIENT_SIG)
	formData.Set("device_country", "us")
	formData.Set("operatorCountry", "us")
	formData.Set("lang", "en")
	formData.Set("sdk_version", "17")

	req, err := http.NewRequest("POST", constants.GOOGLE_AUTH_URL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", constants.GOOGLE_AUTH_USER_AGENT)

	newTransport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}

	httpClient := &http.Client{
		Transport: &newTransport,
		Timeout:   10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respQuery := strings.ReplaceAll(string(bodyBytes), "\n", "&")
	respValues, err := url.ParseQuery(respQuery)
	if err != nil {
		return nil, err
	}

	token := respValues.Get("Auth")

	log.Info().Msg("fetched adm token")

	c.session.AdmSession.AdmToken = token
	err = c.session.SaveSession(c.loggerCtx, constants.DEFAULT_SESSION_FILE)
	if err != nil {
		return &token, err
	}

	return &token, nil
}
