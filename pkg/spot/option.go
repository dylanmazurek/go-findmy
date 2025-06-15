package spot

type Options struct {
	// TODO: Add any options needed for the Spot client.
}

func DefaultOptions() Options {
	defaultOptions := Options{}

	return defaultOptions
}

type Option func(*Options)
