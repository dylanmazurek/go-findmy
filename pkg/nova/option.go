package nova

import (
	"github.com/dylanmazurek/go-findmy/pkg/gpapi"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
)

type Options struct {
	googlePlaySession *gpapi.Session
	notifierSession   *notifier.Session
}

func DefaultOptions() Options {
	defaultOptions := Options{}

	return defaultOptions
}

type Option func(*Options)

func WithGooglePlaySession(s *gpapi.Session) Option {
	return func(o *Options) {
		o.googlePlaySession = s
	}
}

func WithNotifierSession(s *notifier.Session) Option {
	return func(o *Options) {
		o.notifierSession = s
	}
}
