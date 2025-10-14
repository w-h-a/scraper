package scraper

import "context"

type Option func(*Options)

type Options struct {
	Context context.Context
}

func NewOptions(opts ...Option) Options {
	options := Options{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}

type ScrapeOption func(*ScrapeOptions)

type ScrapeOptions struct {
	Context context.Context
}

func NewScrapeOptions(opts ...ScrapeOption) ScrapeOptions {
	options := ScrapeOptions{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}
