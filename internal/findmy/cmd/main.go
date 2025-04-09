package main

import (
	"context"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/findmy"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	ctx := context.Background()

	findmyClient, err := findmy.NewFindMy()
	if err != nil {
		panic(err)
	}

	err = printDevices(ctx, findmyClient)
	if err != nil {
		panic(err)
	}
}

func printDevices(ctx context.Context, findmyClient *findmy.FindMy) error {
	devices := findmyClient.GetDevices(ctx)

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
