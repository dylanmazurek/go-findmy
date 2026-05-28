package gpapi

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GetAASToken(oauthToken string, email string, androidId int64) (*string, error) {
	params := url.Values{}
	params.Set("androidId", fmt.Sprintf("%d", androidId))
	params.Set("lang", "en-US")
	params.Set("google_play_services_version", "212116028")
	params.Set("sdk_version", "28")
	params.Set("device_country", "us")
	params.Set("callerSig", "38918a453d07199354f8b19af05ec6562ced5788")
	params.Set("client_sig", "38918a453d07199354f8b19af05ec6562ced5788")
	params.Set("token_request_options", "CAA4AQ==")
	params.Set("Email", email)
	params.Set("droidguardPeriodicUpdate", "1")
	params.Set("service", "ac2dm")
	params.Set("system_partition", "1")
	params.Set("check_email", "1")
	params.Set("callerPkg", "com.google.android.gms")
	params.Set("get_accountid", "1")
	params.Set("ACCESS_TOKEN", "1")
	params.Set("add_account", "1")
	params.Set("Token", oauthToken)

	req, err := http.NewRequest("POST", "https://android.clients.google.com/auth", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("User-Agent", "Android-Finsky/20.4.13-all%20%5B0%5D%20%5BPR%5D%20313854362 (api=3,versionCode=82041300,sdk=28,device=ASUS_Z01QD_1,hardware=qcom,product=WW_Z01QD,platformVersionRelease=9,model=ASUS_Z01QD,buildId=PKQ1.190101.001,isWideScreen=0)")

	resp, err := c.internalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Token=") {
			token := strings.TrimPrefix(line, "Token=")
			return &token, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("token not found in response, and could not read body: %w", err)
	}

	return nil, fmt.Errorf("token not found in response: %s", string(body))
}
