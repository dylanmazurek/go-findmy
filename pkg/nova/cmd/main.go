package main

import (
	"context"
	"fmt"

	"github.com/dylanmazurek/google-findmy/pkg/nova"
	"github.com/dylanmazurek/google-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/google-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/google-findmy/pkg/shared/session"
	"github.com/markkurossi/tabulate"
)

func main() {
	ctx := context.Background()

	sessionFile := constants.DEFAULT_SESSION_FILE
	session, err := session.New(ctx, &sessionFile)
	if err != nil {
		panic(err)
	}

	clientOps := []nova.Option{
		nova.WithSession(*session),
	}

	novaClient, err := nova.New(ctx, clientOps...)
	if err != nil {
		panic(err)
	}

	//listDevices(novaClient)
	executeAction(novaClient, "670be2bb-0000-2c56-b3c9-089e0832f140")
}

func listDevices(novaClient *nova.Client) error {
	devices, err := novaClient.GetDevices()
	if err != nil {
		return err
	}

	tab := tabulate.New(tabulate.ASCII)
	tab.Header("UUID")
	tab.Header("Name")

	for _, device := range devices.DeviceMetadata {
		var canonicId string
		switch device.GetIdentifierInformation().GetType() {
		case bindings.IdentifierInformationType_IDENTIFIER_ANDROID:
			canonicId = device.GetIdentifierInformation().GetPhoneInformation().GetCanonicIds().GetCanonicId()[0].GetId()
		default:
			canonicId = device.GetIdentifierInformation().GetCanonicIds().GetCanonicId()[0].GetId()
		}

		newRow := tab.Row()
		newRow.Column(canonicId)
		newRow.Column(device.GetUserDefinedDeviceName())
	}

	fmt.Println(tab.String())

	return nil
}

func executeAction(novaClient *nova.Client, canonicId string) error {
	err := novaClient.ExecuteAction(canonicId)

	return err
}
