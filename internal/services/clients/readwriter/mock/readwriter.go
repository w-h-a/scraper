package mock

import (
	"context"

	"github.com/w-h-a/scraper/internal/services/clients/reader"
	"github.com/w-h-a/scraper/internal/services/clients/readwriter"
	"github.com/w-h-a/scraper/internal/services/clients/writer"
)

type mockReadWriter struct {
	options       readwriter.Options
	existingLinks map[string]bool
	readErr       error
	RowsWritten   [][]any
	writeErr      error
}

func (rw *mockReadWriter) ReadExisting(_ context.Context, _ ...reader.ReadExistingOption) (map[string]bool, error) {
	return rw.existingLinks, rw.readErr
}

func (rw *mockReadWriter) WriteBatch(_ context.Context, rows [][]any, opts ...writer.WriteBatchOption) error {
	rw.RowsWritten = rows
	return rw.writeErr
}

func (rw *mockReadWriter) ClearBatch(_ context.Context, opts ...writer.ClearBatchOption) error {
	return nil
}

func NewReadWriter(opts ...readwriter.Option) *mockReadWriter {
	options := readwriter.NewOptions(opts...)

	rw := &mockReadWriter{
		options: options,
	}

	if existing, ok := getExistingLinksFromCtx(options.Context); ok {
		rw.existingLinks = existing
	}

	if err, ok := getReadErrFromCtx(options.Context); ok {
		rw.readErr = err
	}

	if rows, ok := getRowsWrittenFromCtx(options.Context); ok {
		rw.RowsWritten = rows
	}

	if err, ok := getWriteErrFromCtx(options.Context); ok {
		rw.writeErr = err
	}

	return rw
}
