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

	newDevice := models.NewDevice("Test Device", "test serial", "Test Model", "Test Manufacture")
	_, err = publisher.AddDevice(ctx, newDevice)
	if err != nil {
		panic(err)
	}

	// report := shared.LocationReport{
	// 	UniqueId:   nil,
	// 	ReportType: shared.ReportTypeLocation,
	// 	ReportTime: time.Now(),
	// 	Latitude:   37.7749,
	// 	Longitude:  -122.4194,
	// 	Altitude:   0,
	// }

	// _, err = publisher.UpdateTracker(ctx, report)
	// if err != nil {
	// 	panic(err)
	// }
}
