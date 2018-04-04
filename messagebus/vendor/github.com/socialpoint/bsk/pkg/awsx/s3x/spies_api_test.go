package s3x_test

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/socialpoint/bsk/pkg/awsx/s3x"
	"github.com/stretchr/testify/assert"
)

var uri = s3x.NewURI("bucket", "key")

func TestSpyGetObject_GetObject_payload(t *testing.T) {
	a := assert.New(t)

	payload := []byte("payload")
	spy := s3x.SpyGetObjectWithPayload(payload)
	s3xAPI := s3x.New(spy)

	reader, err := s3xAPI.Download(context.Background(), uri)
	a.NoError(err)
	read, err := ioutil.ReadAll(reader)
	a.NoError(err)
	a.Equal(payload, read)
	a.Equal(1, spy.NumCalls())
}

func TestSpyGetObject_GetObject_err(t *testing.T) {
	a := assert.New(t)
	spy := s3x.SpyGetObjectWithError()
	s3xAPI := s3x.New(spy)

	reader, err := s3xAPI.Download(context.Background(), uri)
	a.Error(err)
	a.Nil(reader)
	a.Nil(spy.Payload)
	a.Equal(1, spy.NumCalls())
}

func TestSpyGetObject_GetObject_canceled(t *testing.T) {
	a := assert.New(t)

	payload := []byte("payload")
	spy := s3x.SpyGetObjectWithPayload(payload)
	ctx, cancel := context.WithCancel(context.Background())
	s3xAPI := s3x.New(spy)

	cancel()
	reader, err := s3xAPI.Download(ctx, uri)

	a.Error(err)
	a.Nil(reader)
	a.Equal(1, spy.NumCalls())
}

func TestSpyPutObject_PutObject_payload(t *testing.T) {
	a := assert.New(t)
	spy := &s3x.SpyPutObject{}
	s3xAPI := s3x.New(spy)

	readSeeker := strings.NewReader("payload")
	err := s3xAPI.Upload(context.Background(), uri, readSeeker)

	a.NoError(err)
	a.Equal(string(spy.Payload), "payload")
	a.Equal(1, spy.NumCalls())
}

func TestPutObject_PutObject_err(t *testing.T) {
	a := assert.New(t)
	spy := s3x.SpyPutObjectWithError()
	s3xAPI := s3x.New(spy)

	readSeeker := strings.NewReader("payload")
	err := s3xAPI.Upload(context.Background(), uri, readSeeker)

	a.Error(err)
	a.Nil(spy.Payload)
	a.Equal(1, spy.NumCalls())
}

func TestNopAPI_PutObjectWithContext_canceled(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	spy := &s3x.SpyPutObject{}
	s3xAPI := s3x.New(spy)

	cancel()
	readSeeker := strings.NewReader("payload")
	err := s3xAPI.Upload(ctx, uri, readSeeker)

	a.Error(err)
	a.Nil(spy.Payload)
	a.Equal(1, spy.NumCalls())
}

func TestSpyDeleteObject_DeleteObject(t *testing.T) {
	a := assert.New(t)
	spy := &s3x.SpyDeleteObject{}
	s3xAPI := s3x.New(spy)

	err := s3xAPI.Delete(context.Background(), uri)
	a.NoError(err)
	a.Equal(1, spy.NumCalls())
}

func TestSpyDeleteObject_DeleteObject_err(t *testing.T) {
	a := assert.New(t)
	spy := s3x.SpyDeleteObjectWithError()
	s3xAPI := s3x.New(spy)

	err := s3xAPI.Delete(context.Background(), uri)
	a.Error(err)
	a.Equal(1, spy.NumCalls())
}

func TestSpyDeleteObject_DeleteObject_canceled(t *testing.T) {
	a := assert.New(t)
	spy := s3x.SpyDeleteObjectWithError()
	ctx, cancel := context.WithCancel(context.Background())
	s3xAPI := s3x.New(spy)

	cancel()
	err := s3xAPI.Delete(ctx, uri)
	a.Error(err)
	a.Equal(1, spy.NumCalls())
}

// TestSpies_MockApi verify that spies also implement the MockApi interface
func TestSpies_MockApi(t *testing.T) {
	var mock s3x.MockAPI = &s3x.SpyGetObject{}
	mock = &s3x.SpyPutObject{}
	mock = &s3x.SpyDeleteObject{}
	_ = mock
}
