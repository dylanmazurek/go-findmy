package constants

const (
	CLIENT_NAME = "nova"
)

const (
	API_BASE_URL   = "https://android.googleapis.com/nova"
	API_USER_AGENT = "fmd/20006320; gzip"
	API_LANGUAGE   = "en-US"
)

const (
	CLIENT_APP_GMS = "com.google.android.gms"

	AUTH_OAUTH_SCOPE_BASE = "https://www.googleapis.com/auth/"
	AUTH_CLIENT_SIG       = "38918a453d07199354f8b19af05ec6562ced5788"
	AUTH_CLIENT_SCOPE     = "android_device_manager"
	AUTH_CLIENT_SOURCE    = "android"
)

const (
	PATH_LIST_DEVICES   = "nbe_list_devices"
	PATH_EXECUTE_ACTION = "nbe_execute_action"
)

const (
	LOG_CLIENT_UUID  = "client_uuid"
	LOG_CANONIC_ID   = "canonic_id"
	LOG_DEVICE_ID    = "device_id"
	LOG_REQUEST_UUID = "request_uuid"

	// HTTP request logging
	LOG_HTTP_STATUS = "http_status"
	LOG_HTTP_METHOD = "http_method"
	LOG_HTTP_URL    = "http_url"

	// Message logging
	LOG_MESSAGE_TYPE = "message_type"
)
