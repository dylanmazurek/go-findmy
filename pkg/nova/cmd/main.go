package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/findmy/constants"
	"github.com/dylanmazurek/go-findmy/pkg/gpapi"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/vault"
	"github.com/markkurossi/tabulate"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	log.Trace().Msg("initializing clients")

	secrets, err := getFileSecrets()
	if err != nil {
		log.Error().Err(err).Msg("failed to get file secrets")
		panic(err)
	}

	session := secrets["SESSION"].(map[string]any)
	admSessionObj := session["admSession"].(map[string]any)

	var admSession gpapi.Session = gpapi.Session{
		AndroidId: admSessionObj["androidId"].(float64),
		Email:     admSessionObj["email"].(string),
		AasToken:  admSessionObj["aasToken"].(string),
		AdmToken:  admSessionObj["admToken"].(string),
	}

	clientOps := []nova.Option{
		nova.WithGooglePlaySession(&admSession),
	}

	novaClient, err := nova.NewClient(ctx, clientOps...)
	if err != nil {
		panic(err)
	}

	listDevices(ctx, novaClient)
}

func getVaultSecrets(ctx context.Context) (map[string]any, error) {
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultAppRoleId := os.Getenv("VAULT_APPROLE_ID")
	vaultSecretId := os.Getenv("VAULT_SECRET_ID")

	vaultClient, err := vault.NewClient(ctx, vaultAddr, vaultAppRoleId, vaultSecretId)
	if err != nil {
		return nil, err
	}

	vaultSecret, err := vaultClient.GetSecret(ctx, "kv", "go-findmy")
	if err != nil {
		return nil, err
	}

	if vaultSecret == nil {
		return nil, fmt.Errorf("no vault secret found")
	}

	return vaultSecret, nil
}

func getFileSecrets() (map[string]any, error) {
	var secrets map[string]any

	secretsFile := os.Getenv("SECRETS_FILE")
	if secretsFile == "" {
		secretsFile = fmt.Sprintf("%s/%s", constants.STORAGE_DIR, constants.SECRETS_FILE)
	}

	file, err := os.ReadFile(secretsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	err = json.Unmarshal(file, &secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets file: %w", err)
	}

	return secrets, nil
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
