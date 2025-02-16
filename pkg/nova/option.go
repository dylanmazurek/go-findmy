package nova

import (
	"github.com/dylanmazurek/google-findmy/pkg/shared/session"
)

type Options struct {
	session session.Session
}

func DefaultOptions() Options {
	defaultOptions := Options{}

	return defaultOptions
}

type Option func(*Options)

func WithSession(s session.Session) Option {
	return func(o *Options) {
		o.session = s
	}
}
