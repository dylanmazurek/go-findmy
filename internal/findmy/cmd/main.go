package main

import (
	"context"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/findmy"
	"github.com/dylanmazurek/go-findmy/internal/publisher"
)

func main() {
	//ctx := context.Background()

	// findmyClient, err := findmy.NewFindMy()
	// if err != nil {
	// 	panic(err)
	// }

	// err = initDevices(ctx, findmyClient)
	// if err != nil {
	// 	panic(err)
	// }
}

func initDevices(ctx context.Context, findmyClient *findmy.FindMy) error {
	devices := findmyClient.GetDevices(ctx)

	mqttUrl := os.Getenv("MQTT_URL")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	publisher, err := publisher.NewPublisher(ctx, mqttUrl, mqttUsername, mqttPassword)
	if err != nil {
		return err
	}

	for _, device := range devices {
		_, err = publisher.AddDevice(ctx, device)
		if err != nil {
			return err
		}
	}

	return nil
}
