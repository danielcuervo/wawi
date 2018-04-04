package iox_test

import (
	"testing"

	"github.com/socialpoint/bsk/pkg/awsx/iox"
	"github.com/stretchr/testify/assert"
)

func TestSpyWriteCloser(t *testing.T) {
	a := assert.New(t)
	spy := &iox.SpyWriteCloser{}

	a.Equal("", spy.String())

	n, err := spy.Write([]byte("1234"))
	a.NoError(err)
	a.Equal(4, n)
	a.Equal("1234", spy.String())

	//consecutive writes are concatenated
	n, err = spy.Write([]byte("5"))
	a.NoError(err)
	a.Equal(1, n)
	a.Equal("12345", spy.String())

	//Close() keeps the payload
	err = spy.Close()
	a.NoError(err)
	a.Equal("12345", spy.String())

	//Write after Close() fails but keeps payload
	n, err = spy.Write([]byte("6"))
	a.Error(err)
	a.Equal(0, n)
	a.Equal("12345", spy.String())

}

func TestSpyWriteCloserWithError(t *testing.T) {
	a := assert.New(t)
	spy := iox.ErrWriteCloser{}

	n, err := spy.Write([]byte("1234"))
	a.Error(err)
	a.Equal(0, n)
}
