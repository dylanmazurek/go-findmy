package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dylanmazurek/go-findmy/internal/findmy"
	"github.com/dylanmazurek/go-findmy/internal/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	ctx = logger.InitLogger(ctx)

	// vaultAddr := os.Getenv("VAULT_ADDR")
	// vaultAppRoleId := os.Getenv("VAULT_APP_ROLE_ID")
	// vaultSecretId := os.Getenv("VAULT_SECRET_ID")

	// vaultClient, err := vault.NewClient(ctx, vaultAddr, vaultAppRoleId, vaultSecretId)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("unable to create vault client")
	// }

	findmyClient, err := findmy.NewFindMy()
	if err != nil {
		panic(err)
	}

	err = findmyClient.Start(ctx)
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("waiting for signals")

	<-sigs

	log.Info().Msg("received signal, stopping listener")
}
