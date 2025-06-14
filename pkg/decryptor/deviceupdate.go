package decryptor

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/models"
	"github.com/rs/zerolog/log"
)

func (d *Decryptor) DecryptDeviceUpdate(ctx context.Context, deviceUpdate *bindings.DeviceUpdate) ([]models.LocationReport, error) {
	log := log.Ctx(ctx)

	deviceInformation := deviceUpdate.GetDeviceMetadata().GetInformation()

	encryptedIdentityKey := deviceInformation.GetDeviceRegistration().GetEncryptedUserSecrets().GetEncryptedIdentityKey()
	ownerKey, err := hex.DecodeString(d.OwnerKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode owner key")
		return nil, err
	}

	identityKey, err := decryptEik(ownerKey, encryptedIdentityKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to decrypt eik")
		return nil, err
	}

	locationsProto := deviceInformation.GetLocationInformation().GetReports().GetRecentLocationAndNetworkLocations()

	recentLocation := locationsProto.GetRecentLocation()
	recentLocationTime := locationsProto.GetRecentLocationTimestamp()

	networkLocations := locationsProto.GetNetworkLocations()
	networkLocationsTime := locationsProto.GetNetworkLocationTimestamps()

	if locationsProto.GetRecentLocation() != nil {
		networkLocations = append(networkLocations, recentLocation)
		networkLocationsTime = append(networkLocationsTime, recentLocationTime)
	}

	if len(networkLocations) == 0 {
		log.Trace().Msg("no network locations found in device update")

		return nil, err
	}

	var locations []models.LocationReport
	for i, netLoc := range networkLocations {
		reportTimeInt := networkLocationsTime[i]
		reportTime := time.Unix(int64(reportTimeInt.GetSeconds()), 0)

		var location models.LocationReport
		switch netLoc.GetStatus() {
		case bindings.Status_SEMANTIC:
			loc, err := decryptSemantic(netLoc)
			if err != nil {
				log.Error().Err(err).Msg("failed to decrypt semantic location")
				return nil, err
			}

			location = *loc
		default:
			loc, err := decryptReport(netLoc, identityKey)
			if err != nil {
				log.Error().Err(err).Msg("failed to decrypt report")
				return nil, err
			}

			location = *loc
		}

		location.ReportTime = reportTime
		locations = append(locations, location)
	}

	return locations, nil
}
