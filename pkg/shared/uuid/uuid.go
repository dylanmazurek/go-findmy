package uuid

import (
	"encoding/json"
	"regexp"
	"strings"

	guuid "github.com/google/uuid"
	"github.com/zeebo/xxh3"
)

const (
	ParseTypeUUID     = "raw-uuid"
	ParseTypeComposed = "composed-uuid"
)

type UUID struct {
	uuid guuid.UUID

	ParsedUsing string
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	var uuidString string
	err := json.Unmarshal(data, &uuidString)
	if err != nil {
		return err
	}

	uuid, err := Parse(uuidString)
	if err != nil {
		return err
	}

	u.uuid = uuid.uuid
	u.ParsedUsing = uuid.ParsedUsing

	return nil
}

func (u *UUID) String() (*string, error) {
	if u == nil {
		return nil, ErrUUIDEmpty
	}

	stringUuid := u.uuid.String()
	if stringUuid == "" {
		return nil, ErrUUIDStringFormEmpty
	}

	return &stringUuid, nil
}

func Parse(s string) (*UUID, error) {
	if s == "" {
		return nil, ErrEmptyString
	}

	err := guuid.Validate(s)
	if err == nil {
		uuid, err := guuid.Parse(s)
		if err != nil {
			return nil, err
		}

		guuid := UUID{
			uuid:        uuid,
			ParsedUsing: ParseTypeUUID,
		}

		return &guuid, nil
	}

	cleanStringRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	cleanString := cleanStringRegex.ReplaceAllString(s, "")
	if cleanString == "" {
		return nil, ErrEmptyStringAfterClean
	}

	composedUuidBytes := xxh3.HashString128(cleanString).Bytes()
	uuid, err := guuid.FromBytes(composedUuidBytes[:])
	if err != nil {
		return nil, err
	}

	guuid := UUID{
		uuid:        uuid,
		ParsedUsing: ParseTypeComposed,
	}

	return &guuid, nil
}

func (u *UUID) Short() *string {
	if u == nil {
		return nil
	}

	stringUuid, err := u.String()
	if err != nil {
		return nil
	}

	if stringUuid == nil {
		return nil
	}

	shortUuid := strings.Split(*stringUuid, "-")[4]

	return &shortUuid
}
