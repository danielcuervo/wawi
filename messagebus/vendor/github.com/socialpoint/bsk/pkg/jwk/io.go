package jwk

import (
	"context"
	"time"

	"github.com/socialpoint/bsk/pkg/server"
	"github.com/socialpoint/bsk/pkg/timex"
)

// Reader reads the last n Keys and returns them
type Reader interface {
	Read() ([]*Key, error)
}

// Writer writes Keys
type Writer interface {
	Write(...*Key) error
}

// ReadWriter is the interface that groups the basic Read and Write methods.
type ReadWriter interface {
	Reader
	Writer
}

// LimitReader returns a Reader that reads from r
// but limiting the returned key set to just limit keys
func LimitReader(r Reader, limit int64) ReaderFunc {
	return func() ([]*Key, error) {
		keys, err := r.Read()
		if err != nil {
			return nil, err
		}

		if int64(len(keys)) <= limit {
			return keys, nil
		}

		return keys[0:limit], nil
	}
}

// FilterExpiredReader returns a Reader that reads from r
// but removing expired keys from the resulted slice of keys
func FilterExpiredReader(r Reader, t time.Time) ReaderFunc {
	return func() ([]*Key, error) {
		keys, err := r.Read()
		if err != nil {
			return keys, err
		}

		active := keys[:0]
		for _, k := range keys {
			if t.Before(k.NotAfter) {
				active = append(active, k)
			}
		}

		return active, err
	}
}

// ReaderFunc is an adapter to use simple funcs as a Reader
type ReaderFunc func() ([]*Key, error)

func (f ReaderFunc) Read() ([]*Key, error) {
	return f()
}

// WriterFunc is an adapter to use simple funcs as a Writer
type WriterFunc func(...*Key) error

func (w WriterFunc) Write(keys ...*Key) error {
	return w(keys...)
}

// Copy copies a key set from src Reader to dst Writer
func Copy(dst Writer, src Reader) error {
	keys, err := src.Read()
	if err != nil {
		return err
	}

	return dst.Write(keys...)
}

// CopyEvery copies keys returned by the source reader to the destination writer
// with a period specified by the duration argument.
// Errors are ignored by design, if you want to react to errors, wrap/decorate the
// reader/writer with error handling capabilities.
func CopyEvery(ctx context.Context, d time.Duration, dst Writer, src Reader) {
	go timex.RunInterval(ctx, d, func() {
		_ = Copy(dst, src)
		// By design we drop copy errors on the floor.
		// This runs periodically, so errors should not stop
		// the execution.
		// If stopping the execution is the desired behavior,
		// a reader/writer wrapper could catch errors and cancel
		// the context to stop the periodic execution of this
	})
}

// CopyEveryRunner returns a server.Runner that runs CopyEvery with the provided context
func CopyEveryRunner(d time.Duration, dst Writer, src Reader) server.Runner {
	return server.RunnerFunc(func(ctx context.Context) {
		CopyEvery(ctx, d, dst, src)
	})
}

// TeeReader returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with corresponding writes to w. There is no internal buffering.
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
// This reader is useful when you want to synchronize a source and a destination, but also
// read the synchronized keys.
func TeeReader(r Reader, w Writer) Reader {
	return &teeReader{r, w}
}

type teeReader struct {
	r Reader
	w Writer
}

func (t *teeReader) Read() ([]*Key, error) {
	keys, err := t.r.Read()

	if len(keys) > 0 {
		if err := t.w.Write(keys...); err != nil {
			return nil, err
		}
	}

	return keys, err
}

// MultiReader returns a reader that read from multiple readers and returns an aggregated result
// Read errors are ignored, to handle errors of individual readers, decorate them with this feature.
func MultiReader(readers ...Reader) Reader {
	return ReaderFunc(func() ([]*Key, error) {
		var keys []*Key

		for _, reader := range readers {
			rk, err := reader.Read()
			if err != nil {
				continue
			}

			keys = append(keys, rk...)
		}

		return keys, nil
	})
}

// TODO: Make a version of the multireader that read concurrently from all readers
