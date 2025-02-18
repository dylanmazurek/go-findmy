package main

import (
	"context"
	"encoding/base64"
	"io"
	"os"

	"github.com/dylanmazurek/google-findmy/internal/logger"
	"github.com/dylanmazurek/google-findmy/pkg/decryptor"
	"github.com/dylanmazurek/google-findmy/pkg/nova/models/protos/bindings"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

func main() {
	ctx := context.Background()

	ctx = logger.InitLogger(ctx)
	log := log.Ctx(ctx)

	ownerKey := os.Getenv("OWNER_KEY")
	newDecryptor := decryptor.NewDecryptor(ownerKey)

	fcmPayloadEncodedFile, err := os.OpenFile("message.txt", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	log.Info().Msg("Opened message.txt")

	fcmPayloadEncoded, err := io.ReadAll(fcmPayloadEncodedFile)
	if err != nil {
		panic(err)
	}

	fcmPayload, err := base64.RawStdEncoding.DecodeString(string(fcmPayloadEncoded))
	if err != nil {
		panic(err)
	}

	var deviceUpdate bindings.DeviceUpdate
	err = proto.Unmarshal([]byte(fcmPayload), &deviceUpdate)
	if err != nil {
		panic(err)
	}

	newDecryptor.DecryptLocations(ctx, &deviceUpdate)
}
