package mock

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/w-h-a/scraper/internal/services/clients/scraper"
)

type mockScraper struct {
	options      scraper.Options
	feedToReturn *gofeed.Feed
	errToReturn  error
}

func (s *mockScraper) Scrape(_ context.Context, _ string, _ ...scraper.ScrapeOption) (*gofeed.Feed, error) {
	return s.feedToReturn, s.errToReturn
}

func NewScraper(opts ...scraper.Option) *mockScraper {
	options := scraper.NewOptions(opts...)

	s := &mockScraper{
		options: options,
	}

	if fd, ok := getFeedFromCtx(options.Context); ok {
		s.feedToReturn = fd
	}

	if err, ok := getErrFromCtx(options.Context); ok {
		s.errToReturn = err
	}

	return s
}
