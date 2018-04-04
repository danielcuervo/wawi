package s3x

import (
	"context"
	"io"
)

// API simplifies s3iface.S3API with simpler functions: only 1 return value when possible, use URI,...
// Methods must be safe to use concurrently.
type API interface {
	Delete(ctx context.Context, uri *URI) error
	Download(ctx context.Context, uri *URI) (io.ReadCloser, error)
	Upload(ctx context.Context, uri *URI, input io.ReadSeeker) error
}
