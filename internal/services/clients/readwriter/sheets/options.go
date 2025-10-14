package sheets

import (
	"context"

	"github.com/w-h-a/scraper/internal/services/clients/readwriter"
)

type serviceAccountKeyPathKey struct{}

func WithServiceAccountKeyPath(path string) readwriter.Option {
	return func(o *readwriter.Options) {
		o.Context = context.WithValue(o.Context, serviceAccountKeyPathKey{}, path)
	}
}

func getServiceAccountKeyPathFromCtx(context context.Context) (string, bool) {
	path, ok := context.Value(serviceAccountKeyPathKey{}).(string)
	return path, ok
}
