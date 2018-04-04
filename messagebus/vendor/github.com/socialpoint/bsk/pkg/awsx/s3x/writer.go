package s3x

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// S3Writer creates an io.WriteCloser suitable for writing to an S3 object
// TODO allocates a buffer of at least 5MB.
func S3Writer(ctx context.Context, uploader s3manageriface.UploaderAPI, bucket string, key string) io.WriteCloser {
	reader, writer := io.Pipe()

	w := &s3writer{
		reader:   reader,
		writer:   writer,
		uploader: uploader,

		errs: make(chan error, 1),
		ctx:  ctx,

		bucket: bucket,
		key:    key,
	}

	go func() {
		_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   reader,
		})
		//don't close the reader here. Otherwise, Write() might exit before reading errs
		w.errs <- w.decorate(err)
		_ = reader.Close()
	}()

	return w
}

type s3writer struct {
	reader io.ReadCloser
	writer io.WriteCloser

	uploader s3manageriface.UploaderAPI

	errs chan error
	ctx  context.Context

	bucket string
	key    string
}

// Write writes what uploader has written into the pipe so far
func (w *s3writer) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil && err != io.ErrClosedPipe {
		return n, err
	}
	select {
	case err = <-w.errs:
		close(w.errs)
	default:
	}
	return n, err
}

// Close will return nil if Write has already returned a non-nil error
func (w *s3writer) Close() error {
	_ = w.writer.Close()
	err, ok := <-w.errs
	if ok {
		return err
	}
	return nil
}

func (w *s3writer) decorate(err error) error {
	return decorateError(err, "s3writer", NewURI(w.bucket, w.key))
}
