package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dylanmazurek/go-findmy/internal/findmy"
	"github.com/dylanmazurek/go-findmy/internal/logger"
	"github.com/dylanmazurek/go-findmy/pkg/shared/vault"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	ctx = logger.InitLogger(ctx)

	secrets, err := getSecrets(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get secrets")
		os.Exit(1)
	}

	findmyService, err := findmy.NewService(ctx, secrets)
	if err != nil {
		panic(err)
	}

	err = findmyService.Start(ctx)
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("go-findmy service running")

	<-sigs

	log.Info().Msg("received terminate signal, stopping listener")
}

func getSecrets(ctx context.Context) (map[string]any, error) {
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
