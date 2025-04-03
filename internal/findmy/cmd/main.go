package main

import (
	"context"

	"github.com/dylanmazurek/go-findmy/internal/findmy"
)

func main() {
	ctx := context.Background()

	findmyClient, err := findmy.NewFindMy()
	if err != nil {
		panic(err)
	}

	err = initDevices(ctx, findmyClient)
	if err != nil {
		panic(err)
	}
}

func initDevices(ctx context.Context, findmyClient *findmy.FindMy) error {
	devices := findmyClient.GetDevices(ctx)

	for _, device := range devices {
		err := findmyClient.PublishDevice(ctx, device)
		if err != nil {
			return err
		}
	}

	return nil
}
