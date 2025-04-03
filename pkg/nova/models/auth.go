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

	ReceivedAt time.Time
}

func (a *Auth) Expiry() *time.Time {
	if a == nil {
		return nil
	}

	if a.ReceivedAt.IsZero() {
		a.ReceivedAt = time.Now()
	}

	if a.ExpiryInt > 0 {
		expiryTime := time.Unix(int64(a.ExpiryInt), 0)
		return &expiryTime
	}

	if a.ExpiresInSecInt > 0 {
		expiryTime := a.ReceivedAt.Add(time.Duration(a.ExpiresInSecInt) * time.Second)
		return &expiryTime
	}

	pastTime := time.Now().Add(-1 * time.Minute)
	return &pastTime
}
