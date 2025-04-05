package nova

import "errors"

// auth
var (
	ErrFailedToInitializeClients = errors.New("failed to initialize clients")
	ErrFailedToPrintDevices      = errors.New("failed to print devices")
	ErrFailedToExecuteAction     = errors.New("failed to execute action")
	ErrFailedToGetAdmToken       = errors.New("failed to get adm token")
	ErrTokenExpired              = errors.New("token expired")
)

// response
var (
	ErrResponseNotProtoMessage = errors.New("response is not a proto.message")
)
