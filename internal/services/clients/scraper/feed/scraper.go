package feed

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/w-h-a/scraper/internal/services/clients/scraper"
)

type feedScraper struct {
	options scraper.Options
	parser  *gofeed.Parser
}

func (s *feedScraper) Scrape(ctx context.Context, url string, _ ...scraper.ScrapeOption) (*gofeed.Feed, error) {
	return s.parser.ParseURLWithContext(url, ctx)
}

func NewScraper(opts ...scraper.Option) scraper.Scraper {
	options := scraper.NewOptions(opts...)

	s := &feedScraper{
		options: options,
		parser:  gofeed.NewParser(),
	}

	return s
}
