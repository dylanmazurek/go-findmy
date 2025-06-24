package constants

import "regexp"

// Common constants
const (
	GOOGLE_AUTH_URL        = "https://android.clients.google.com/auth"
	GOOGLE_AUTH_USER_AGENT = "GoogleAuth/1.4"
	GOOGLE_TOKEN_INFO_URL  = "https://www.googleapis.com/oauth2/v3/tokeninfo"

	ADM_APP_ID = "com.google.android.apps.adm"
)

// Constants for the session
var (
	DEFAULT_SESSION_FILE = ".storage/session.json"
)

var (
	UNIQUE_ID_REGEX = regexp.MustCompile("[^a-zA-Z0-9]+")
)
