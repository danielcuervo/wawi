package iox

import (
	"bytes"
	"fmt"
	"io"

	"github.com/socialpoint/bsk/pkg/testx"
)

//ErrReadCloser is a ReadCloser stub
type ErrReadCloser struct {
}

// Read always fails
func (e ErrReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("ErrReadCloser")
}

// Close always succeeds
func (e ErrReadCloser) Close() error {
	return nil
}

type dummyReadCloser struct {
	testx.DummyCloser
	*bytes.Buffer
}

// DummyReadCloser creates a ReadCloser with a specified buffer
func DummyReadCloser(buf []byte) io.ReadCloser {
	return &dummyReadCloser{Buffer: bytes.NewBuffer(buf)}
}

// Read will provide the buffer until EOF or closed
func (r *dummyReadCloser) Read(p []byte) (n int, err error) {
	if r.IsClosed() {
		return 0, fmt.Errorf("DummyReadCloser is closed")
	}
	return r.Buffer.Read(p)
}
