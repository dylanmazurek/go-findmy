package models

import (
	"strconv"
	"time"
)

type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalText(text []byte) error {
	timeInt64, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}

	timeUnix := time.Unix(timeInt64, 0).UTC()

	*u = UnixTime{timeUnix}

	return nil
}

type Auth struct {
	IssueAdvice          string `schema:"issueAdvice"`
	StoreConsentRemotely bool   `schema:"storeConsentRemotely"`
	IsTokenSnowballed    bool   `schema:"isTokenSnowballed"`
	GrantedScopes        string `schema:"grantedScopes"`
	Token                string `schema:"Auth"`

	ExpiresAt UnixTime `schema:"Expiry"`
}

func (a *Auth) IsValid() bool {
	if a == nil {
		return false
	}

	if a.Token == "" {
		return false
	}

	if a.ExpiresAt.IsZero() {
		return false
	}

	expiresIn := time.Until(a.ExpiresAt.Time)
	hasExpired := expiresIn < 0

	return !hasExpired
}
