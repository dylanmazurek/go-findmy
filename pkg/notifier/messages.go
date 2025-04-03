package notifier

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/dylanmazurek/go-findmy/internal"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	shared "github.com/dylanmazurek/go-findmy/pkg/shared/models"
	"github.com/rs/zerolog/log"
)

func (n *Client) handleReport(ctx context.Context, deviceUpdate *bindings.DeviceUpdate, loc shared.LocationReport) (*shared.LocationReport, error) {
	log := log.Ctx(ctx)

	uniqueId, err := internal.FormatUniqueId(deviceUpdate.DeviceMetadata)
	if errors.Is(err, errors.New("multiple canonic ids found")) {
		log.Warn().Err(err).Msg("multiple canonic ids found, using first one")
	}

	var newLocationReport shared.LocationReport = loc
	newLocationReport.UniqueId = uniqueId
	newLocationReport.ReportTime = loc.ReportTime

	isSemanticLocation := (loc.SemanticName != nil)

	if isSemanticLocation {
		err := processSemanticLocation(&newLocationReport)
		if err != nil {
			log.Error().
				Str("semantic_name", *loc.SemanticName).
				Err(err).Msg("failed to process semantic location")

			return nil, err
		}

		return &newLocationReport, nil
	}

	newLocationReport.ReportType = shared.ReportTypeLocation
	newLocationReport.Latitude = loc.Latitude
	newLocationReport.Longitude = loc.Longitude
	newLocationReport.Altitude = loc.Altitude

	accuracy := deviceUpdate.GetDeviceMetadata().
		GetInformation().GetLocationInformation().
		GetReports().RecentLocationAndNetworkLocations.
		GetRecentLocation().GetGeoLocation().GetAccuracy()

	newLocationReport.Accuracy = float64(accuracy)

	return &newLocationReport, nil
}

func processSemanticLocation(locationReport *shared.LocationReport) error {
	semanticLocationName := *locationReport.SemanticName
	if semanticLocationName == "" {
		err := fmt.Errorf("failed to find semantic location name")
		return err
	}

	semanticLocationIdx := slices.IndexFunc(semanticLocations, func(i shared.SemanticLocation) bool {
		hasName := slices.Contains(i.Names, semanticLocationName)
		return hasName
	})

	if semanticLocationIdx == -1 {
		err := fmt.Errorf("failed to find semantic location")

		return err
	}

	semanticLocation := semanticLocations[semanticLocationIdx]

	locationReport.ReportType = shared.ReportTypeSemantic
	locationReport.Latitude = semanticLocation.Latitude
	locationReport.Longitude = semanticLocation.Longitude
	locationReport.Accuracy = 4
	locationReport.Altitude = 0
	locationReport.SemanticName = &semanticLocation.Names[0]

	return nil
}
