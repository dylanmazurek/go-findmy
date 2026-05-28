package main

import (
	"os"
	"strconv"

	"github.com/dylanmazurek/go-findmy/pkg/gpapi"
)

func main() {
	newClient := gpapi.NewClient()

	oauthToken := os.Getenv("GO_FINDMY_OAUTH_TOKEN")
	if oauthToken == "" {
		panic("GO_FINDMY_OAUTH_TOKEN environment variable is not set")
	}

	email := os.Getenv("GO_FINDMY_EMAIL")
	if email == "" {
		panic("GO_FINDMY_EMAIL environment variable is not set")
	}

	androidIdStr := os.Getenv("GO_FINDMY_ANDROID_ID")
	if androidIdStr == "" {
		panic("GO_FINDMY_ANDROID_ID environment variable is not set")
	}

	androidId, err := strconv.ParseInt(androidIdStr, 10, 64)
	if err != nil {
		panic("GO_FINDMY_ANDROID_ID must be a valid integer")
	}

	aasToken, err := newClient.GetAASToken(oauthToken, email, androidId)
	if err != nil {
		panic("Failed to get AAS token: " + err.Error())
	}

	if aasToken == nil {
		panic("AAS token is nil")
	}

	println("AAS Token:", *aasToken)
	println("AAS Token Length:", len(*aasToken))
}
