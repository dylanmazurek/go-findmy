package main

import (
	"context"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/findmy"
	"github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	findmyService, err := findmy.NewService(ctx)
	if err != nil {
		panic(err)
	}

	devices, err := findmyService.GetDevices(ctx)
	if err != nil {
		panic(err)
	}

	err = printDevices(devices)
	if err != nil {
		panic(err)
	}

	err = publishDevices(ctx, findmyService, devices)
	if err != nil {
		panic(err)
	}
}

func printDevices(devices []models.Device) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	t.AppendHeader(table.Row{"ID", "Name", "Model"})
	for _, device := range devices {
		t.AppendRow(table.Row{
			device.UniqueId,
			device.DeviceInfo.Name,
			device.DeviceInfo.Model,
		})
	}

	t.Render()

	return nil
}

func publishDevices(ctx context.Context, findmyService *findmy.Service, devices []models.Device) error {
	log := log.Ctx(ctx)

	log.Info().Msg("publishing devices to mqtt")

	if len(devices) == 0 {
		return nil
	}

	for _, device := range devices {
		err := findmyService.PublishDevice(ctx, device)
		if err != nil {
			return err
		}
	}

	return nil
}
