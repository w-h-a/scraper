package reader

import "context"

type ReadExistingOption func(*ReadExistingOptions)

type ReadExistingOptions struct {
	Query   string
	Context context.Context
}

func ReadExistingWithQuery(query string) ReadExistingOption {
	return func(reo *ReadExistingOptions) {
		reo.Query = query
	}
}

func NewReadExistingOptions(opts ...ReadExistingOption) ReadExistingOptions {
	options := ReadExistingOptions{
		Context: context.Background(),
	}

	for _, fn := range opts {
		fn(&options)
	}

	return options
}
