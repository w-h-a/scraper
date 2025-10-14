package mock

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/w-h-a/scraper/internal/services/clients/scraper"
)

type feedKey struct{}
type errKey struct{}

func WithFeed(fd *gofeed.Feed) scraper.Option {
	return func(o *scraper.Options) {
		o.Context = context.WithValue(o.Context, feedKey{}, fd)
	}
}

func getFeedFromCtx(ctx context.Context) (*gofeed.Feed, bool) {
	fd, ok := ctx.Value(feedKey{}).(*gofeed.Feed)
	return fd, ok
}

func WithErr(err error) scraper.Option {
	return func(o *scraper.Options) {
		o.Context = context.WithValue(o.Context, errKey{}, err)
	}
}

func getErrFromCtx(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(errKey{}).(error)
	return err, ok
}
