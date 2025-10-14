package scraper

import (
	"context"

	"github.com/mmcdole/gofeed"
)

type ScraperType string

const (
	Mock ScraperType = "mock"
	Feed ScraperType = "feed"
)

var (
	ScraperTypes = map[string]ScraperType{
		"mock": Mock,
		"feed": Feed,
	}
)

type Scraper interface {
	Scrape(ctx context.Context, url string, opts ...ScrapeOption) (*gofeed.Feed, error)
}
