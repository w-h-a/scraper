package writer

import "context"

type WriteBatchOption func(*WriteBatchOptions)

type WriteBatchOptions struct {
	Context context.Context
}

func NewWriteBatchOption(opts ...WriteBatchOption) WriteBatchOptions {
	options := WriteBatchOptions{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}

type ClearBatchOption func(*ClearBatchOptions)

type ClearBatchOptions struct {
	Context context.Context
}

func NewClearBatchOptions(opts ...ClearBatchOption) ClearBatchOptions {
	options := ClearBatchOptions{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}
