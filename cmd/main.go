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
