package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/vault"
	"github.com/markkurossi/tabulate"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	log.Trace().Msg("initializing clients")
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultAppRoleId := os.Getenv("VAULT_APPROLE_ID")
	vaultSecretId := os.Getenv("VAULT_SECRET_ID")

	vaultClient, err := vault.NewClient(ctx, vaultAddr, vaultAppRoleId, vaultSecretId)
	if err != nil {
		panic(err)
	}

	vaultSecret, err := vaultClient.GetSecret(ctx, "kv", "go-findmy")
	if err != nil {
		panic(err)
	}

	sessionIrf, ok := vaultSecret["SESSION"].(map[string]interface{})
	if !ok {
		err := fmt.Errorf("SESSION not found in vault secret")
		panic(err)
	}

	sessionBytes, err := json.Marshal(sessionIrf)
	if err != nil {
		panic(err)
	}

	var session *notifier.Session
	err = json.Unmarshal([]byte(sessionBytes), &session)
	if err != nil {
		panic(err)
	}

	clientOps := []nova.Option{
		nova.WithNotifierSession(session),
	}

	novaClient, err := nova.NewClient(ctx, clientOps...)
	if err != nil {
		panic(err)
	}

	listDevices(ctx, novaClient)
}

func listDevices(ctx context.Context, novaClient *nova.Client) error {
	log := log.Ctx(ctx)

	log.Info().Msg("getting devices")

	devices, err := novaClient.GetDevices(ctx)
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
			phoneInfo := device.GetIdentifierInformation().GetPhoneInformation()
			if phoneInfo == nil {
				continue
			}

			canonicIds := phoneInfo.GetCanonicIds()
			if canonicIds == nil {
				continue
			}

			canonicId = canonicIds.GetCanonicId()[0].GetId()
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
