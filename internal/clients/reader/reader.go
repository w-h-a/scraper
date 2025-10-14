package reader

import "context"

type Reader interface {
	ReadExisting(ctx context.Context, opts ...ReadExistingOption) (map[string]bool, error)
}
