package iox

import (
	"bytes"
	"fmt"

	"github.com/socialpoint/bsk/pkg/testx"
)

// ErrWriteCloser is a WriteCloser which always fails
type ErrWriteCloser struct {
}

// Write always fails
func (e ErrWriteCloser) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("ErrWriteCloser")
}

// Close always succeeds
func (e ErrWriteCloser) Close() error {
	return nil
}

// SpyWriteCloser spies the written data
type SpyWriteCloser struct {
	bytes.Buffer
	testx.DummyCloser
}

// Write will append the payload into the buffer unless it's closed
func (r *SpyWriteCloser) Write(p []byte) (n int, err error) {
	if r.IsClosed() {
		return 0, fmt.Errorf("SpyWriteCloser is closed")
	}
	return r.Buffer.Write(p)
}
