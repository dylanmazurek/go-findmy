package findmy

import (
	"context"

	"github.com/dylanmazurek/go-findmy/internal"
	pubModels "github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/rs/zerolog/log"
)

func (s *Service) GetDevices(ctx context.Context) ([]pubModels.Device, error) {
	devices, err := s.novaClient.GetDevices(ctx)
	if err != nil {
		return nil, err
	}

	var pubDevices []pubModels.Device
	for _, device := range devices.DeviceMetadata {
		deviceName := device.GetUserDefinedDeviceName()

		deviceType, err := internal.GetDeviceType(device)
		if err != nil {
			log.Error().Err(err).Msg("failed to get device type")

			continue
		}

		var model, manufacturer string

		switch deviceType {
		case bindings.IdentifierInformationType_IDENTIFIER_SPOT:
			model = device.GetInformation().GetDeviceRegistration().GetModel()
			manufacturer = device.GetInformation().GetDeviceRegistration().GetManufacturer()
		}

		serial, err := internal.FormatUniqueId(device)
		if err != nil {
			continue
		}

		newPubDevice := pubModels.NewDevice(deviceName, *serial, model, manufacturer)
		pubDevices = append(pubDevices, newPubDevice)
	}

	return pubDevices, nil
}

func (s *Service) PublishDevice(ctx context.Context, device pubModels.Device) error {
	log := log.Ctx(ctx)

	resp, err := s.publisherClient.AddDevice(ctx, device)
	if err != nil {
		log.Error().Err(err).Msg("failed to publish device")
		return err
	}

	if resp == nil {
		log.Error().Msg("response is nil")
	}

	return nil
}
