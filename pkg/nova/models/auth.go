package models

import "time"

type Auth struct {
	IssueAdvice          string `form:"issueAdvice"`
	StoreConsentRemotely bool   `form:"storeConsentRemotely"`
	IsTokenSnowballed    bool   `form:"isTokenSnowballed"`
	GrantedScopes        string `form:"grantedScopes"`
	Token                string `form:"Auth"`

	ExpiryInt       int `form:"Expiry"`
	ExpiresInSecInt int `form:"ExpiresInDurationSec"`
}

func (a *Auth) Expiry() *time.Time {
	if a == nil {
		return nil
	}

	if a.ExpiryInt == 0 {
		return nil
	}

	expiryTime := time.Unix(int64(a.ExpiryInt), 0)
	return &expiryTime
}
