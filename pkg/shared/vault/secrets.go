package vault

import (
	"context"
	"fmt"
)

func (c *Client) GetSecret(ctx context.Context, mountPath string, secretPath string) (map[string]interface{}, error) {
	secret, err := c.internalClient.KVv2(mountPath).Get(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	return secret.Data, nil
}

func (c *Client) InsertSecret(ctx context.Context, mountPath string, secretPath string, data map[string]interface{}) error {
	_, err := c.internalClient.KVv2(mountPath).Put(ctx, secretPath, data)
	if err != nil {
		return fmt.Errorf("unable to insert secret: %w", err)
	}

	return nil
}
