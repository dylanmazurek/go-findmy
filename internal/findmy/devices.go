package findmy

import (
	"context"

	"github.com/dylanmazurek/go-findmy/internal"
	pubModels "github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/rs/zerolog/log"
)

func (f *FindMy) GetDevices(ctx context.Context) []pubModels.Device {
	log := log.Ctx(ctx)

	devices, err := f.novaClient.GetDevices(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get devices")
	}

	var pubDevices []pubModels.Device
	for _, device := range devices.DeviceMetadata {
		deviceName := device.GetUserDefinedDeviceName()
		model := device.GetInformation().GetDeviceRegistration().GetModel()
		manufacturer := device.GetInformation().GetDeviceRegistration().GetManufacturer()

		serial, err := internal.FormatUniqueId(device)
		if err != nil {
			continue
		}

		newPubDevice := pubModels.NewDevice(deviceName, *serial, model, manufacturer)
		pubDevices = append(pubDevices, newPubDevice)
	}

	return pubDevices
}

func (f *FindMy) PublishDevice(ctx context.Context, device pubModels.Device) error {
	log := log.Ctx(ctx)

	// Publish the device to the publisher client
	resp, err := f.publisherClient.AddDevice(ctx, device)
	if err != nil {
		log.Error().Err(err).Msg("failed to publish device")
		return err
	}

	// Check if the response is nil
	if resp == nil {
		log.Error().Msg("response is nil")
	}

	return nil
}
