package main

import (
	"context"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/internal/publisher/models"
)

func main() {
	ctx := context.Background()

	mqttUrl := os.Getenv("MQTT_URL")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	publisher, err := publisher.NewPublisher(ctx, mqttUrl, mqttUsername, mqttPassword)
	if err != nil {
		panic(err)
	}

	// newDevice := models.NewDevice("Test Device", "test serial", "Test Model", "Test Manufacture")
	// _, err = publisher.AddDevice(ctx, newDevice)
	// if err != nil {
	// 	panic(err)
	// }

	updateRecord := models.Attributes{
		Latitude:    32.87336,
		Longitude:   -117.22743,
		GpsAccuracy: 1.2,
	}

	_, err = publisher.UpdateTracker(ctx, "", updateRecord)
	if err != nil {
		panic(err)
	}
}
