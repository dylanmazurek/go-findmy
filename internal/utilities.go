package internal

import (
	"errors"
	"strings"

	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
)

func GetDeviceType(deviceMetadata *bindings.DeviceMetadata) (bindings.IdentifierInformationType, error) {
	identifierInformation := deviceMetadata.GetIdentifierInformation()
	if identifierInformation == nil {
		return bindings.IdentifierInformationType_IDENTIFIER_UNKNOWN, errors.New("identifier information is nil")
	}

	switch identifierInformation.Type {
	case bindings.IdentifierInformationType_IDENTIFIER_SPOT:
		return bindings.IdentifierInformationType_IDENTIFIER_SPOT, nil
	case bindings.IdentifierInformationType_IDENTIFIER_ANDROID:
		return bindings.IdentifierInformationType_IDENTIFIER_ANDROID, nil
	}

	return bindings.IdentifierInformationType_IDENTIFIER_UNKNOWN, errors.New("unknown identifier information type")
}

func FormatUniqueId(deviceMetadata *bindings.DeviceMetadata) (*string, error) {
	canonicIds := deviceMetadata.GetIdentifierInformation().GetCanonicIds().GetCanonicId()
	if len(canonicIds) == 0 {
		canonicIds = deviceMetadata.GetIdentifierInformation().GetPhoneInformation().GetCanonicIds().GetCanonicId()
		if len(canonicIds) == 0 {
			return nil, errors.New("no canonic ids found")
		}
	}

	firstCanonicId := canonicIds[0].GetId()

	uniqueId := strings.ToLower(firstCanonicId)

	if len(canonicIds) > 1 {
		err := errors.New("multiple canonic ids found")
		return &uniqueId, err
	}

	return &uniqueId, nil
}
