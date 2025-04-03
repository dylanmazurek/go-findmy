package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

func InitLogger(ctx context.Context) context.Context {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "Mon, 02 Jan 2006 03:04:05"}

	logLevelEnv, hasLogLevel := os.LookupEnv("LOG_LEVEL")
	if !hasLogLevel {
		logLevelEnv = "info"
	}

	logLevel, err := zerolog.ParseLevel(logLevelEnv)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	log := zerolog.New(output).With().Timestamp().Logger()

	ctx = log.WithContext(ctx)

	return ctx
}
