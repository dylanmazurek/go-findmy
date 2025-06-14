package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dylanmazurek/go-findmy/internal/logger"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	ctx = logger.InitLogger(ctx)

	sessionFile := constants.DEFAULT_SESSION_FILE
	session, err := notifier.NewSession(ctx, &sessionFile)
	if err != nil {
		panic(err)
	}

	n, err := notifier.NewClient(ctx, session, nil, nil)
	if err != nil {
		panic(err)
	}

	err = n.StartListening(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to start listening")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("notifier running")

	<-sigs

	log.Info().Msg("received terminate signal, stopping listener")

}
