package uuid

import "errors"

var (
	ErrUUIDEmpty             = errors.New("uuid is empty")
	ErrUUIDStringFormEmpty   = errors.New("uuid string form is empty")
	ErrEmptyString           = errors.New("empty input string")
	ErrEmptyStringAfterClean = errors.New("empty input string after clean")
)
