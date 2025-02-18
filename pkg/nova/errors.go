package nova

import "errors"

var (
	ErrFailedToInitializeClients = errors.New("failed to initialize clients")
	ErrFailedToPrintDevices      = errors.New("failed to print devices")
	ErrFailedToExecuteAction     = errors.New("failed to execute action")
	ErrFailedToGetAdmToken       = errors.New("failed to get adm token")
)

var (
	ErrTokenExpired = errors.New("token expired")
)
