package writer

import "context"

type Writer interface {
	WriteBatch(ctx context.Context, rows [][]any, opts ...WriteBatchOption) error
	ClearBatch(ctx context.Context, opts ...ClearBatchOption) error
}
