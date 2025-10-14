package mock

import (
	"context"

	"github.com/w-h-a/scraper/internal/clients/readwriter"
)

type existingLinksKey struct{}
type readErrKey struct{}
type rowsWrittenKey struct{}
type writeErrKey struct{}

func WithExistingLinksKey(existing map[string]bool) readwriter.Option {
	return func(o *readwriter.Options) {
		o.Context = context.WithValue(o.Context, existingLinksKey{}, existing)
	}
}

func getExistingLinksFromCtx(ctx context.Context) (map[string]bool, bool) {
	existing, ok := ctx.Value(existingLinksKey{}).(map[string]bool)
	return existing, ok
}

func WithReadErr(err error) readwriter.Option {
	return func(o *readwriter.Options) {
		o.Context = context.WithValue(o.Context, readErrKey{}, err)
	}
}

func getReadErrFromCtx(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(readErrKey{}).(error)
	return err, ok
}

func WithRowsWritten(rows [][]any) readwriter.Option {
	return func(o *readwriter.Options) {
		o.Context = context.WithValue(o.Context, rowsWrittenKey{}, rows)
	}
}

func getRowsWrittenFromCtx(ctx context.Context) ([][]any, bool) {
	rows, ok := ctx.Value(rowsWrittenKey{}).([][]any)
	return rows, ok
}

func WithWriteErr(err error) readwriter.Option {
	return func(o *readwriter.Options) {
		o.Context = context.WithValue(o.Context, writeErrKey{}, err)
	}
}

func getWriteErrFromCtx(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(writeErrKey{}).(error)
	return err, ok
}
