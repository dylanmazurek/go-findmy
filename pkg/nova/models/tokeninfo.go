package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type TokenInfo struct {
	Azp   string `json:"azp"`
	Aud   string `json:"aud"`
	Scope string `json:"scope"`

	AccessType AccessType `json:"access_type"`
	Expiry     time.Time  `json:"-"`
}

func (t *TokenInfo) UnmarshalJSON(data []byte) error {
	type Alias TokenInfo
	aux := &struct {
		ExpiryDate string `json:"exp"`
		ExpiresIn  string `json:"expires_in"`
		AccessType string `json:"access_type"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	expiryDateInt, err := strconv.ParseInt(aux.ExpiryDate, 10, 64)
	if err != nil {
		return err
	}

	t.Expiry = time.Unix(expiryDateInt, 0)

	t.AccessType = AccessType(aux.AccessType)

	return nil
}

func (t *TokenInfo) IsValid() bool {
	hasExpired := t.Expiry.Before(time.Now())
	return !hasExpired
}

type AccessType string

const (
	Offline AccessType = "offline"
)
