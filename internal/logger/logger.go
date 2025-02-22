package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func InitLogger(ctx context.Context) context.Context {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp}

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log := zerolog.New(output).With().Timestamp().Logger()

	ctx = log.WithContext(ctx)

	return ctx
}
