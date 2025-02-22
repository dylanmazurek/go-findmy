package vault

import (
	"context"
	"fmt"

	vaultApi "github.com/hashicorp/vault/api"
	vaultAuth "github.com/hashicorp/vault/api/auth/approle"
)

type Client struct {
	internalClient *vaultApi.Client
}

func NewClient(ctx context.Context, addr string, appRoleId string, secretId string) (*Client, error) {
	config := vaultApi.DefaultConfig()
	config.Address = addr

	client, err := vaultApi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	appRoleAuth, err := vaultAuth.NewAppRoleAuth(
		appRoleId,
		&vaultAuth.SecretID{FromString: secretId},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AppRole auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(ctx, appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}

	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	newClient := &Client{
		internalClient: client,
	}

	return newClient, nil
}
