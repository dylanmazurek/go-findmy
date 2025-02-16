package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dylanmazurek/google-findmy/internal/logger"
	"github.com/dylanmazurek/google-findmy/pkg/notifier"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	ctx = logger.InitLogger(ctx)

	n := notifier.NewNotifier(ctx)

	err := n.StartListening()
	if err != nil {
		log.Error().Err(err).Msg("failed to start listening")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("waiting for signals")

	<-sigs

	log.Info().Msg("received signal, stopping listener")

}
