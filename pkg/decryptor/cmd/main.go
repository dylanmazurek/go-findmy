package main

import (
	"context"
	"encoding/base64"
	"io"
	"os"

	"github.com/dylanmazurek/go-findmy/internal/logger"
	"github.com/dylanmazurek/go-findmy/pkg/decryptor"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/constants"
	"github.com/dylanmazurek/go-findmy/pkg/shared/session"
	"google.golang.org/protobuf/proto"
)

func main() {
	ctx := context.Background()

	ctx = logger.InitLogger(ctx)

	sessionFile := constants.DEFAULT_SESSION_FILE
	session, err := session.New(ctx, &sessionFile)
	if err != nil {
		panic(err)
	}

	newDecryptor, err := decryptor.NewDecryptor(session.OwnerKey)
	if err != nil {
		panic(err)
	}

	fcmPayloadEncodedFile, err := os.OpenFile("message.txt", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

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

	locationReports, err := newDecryptor.DecryptDeviceUpdate(ctx, &deviceUpdate)
	if err != nil {
		panic(err)
	}

	_ = locationReports
}
