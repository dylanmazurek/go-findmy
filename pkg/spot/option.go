package spot

import (
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
)

type Options struct {
	notifierSession notifier.Session
}

func DefaultOptions() Options {
	defaultOptions := Options{}

	return defaultOptions
}

type Option func(*Options)

func WithSession(s notifier.Session) Option {
	return func(o *Options) {
		o.notifierSession = s
	}
}
