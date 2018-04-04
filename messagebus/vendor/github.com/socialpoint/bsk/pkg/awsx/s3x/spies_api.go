package s3x

import (
	"bytes"
	"errors"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// MockAPI exposes mockS3's methods
type MockAPI interface {
	s3iface.S3API
	IsErr() bool
	NumCalls() int
}

// mockS3 is a s3iface.S3API which spies the calls to a single S3 method and NoOps the other ones
// Only ...WithContext methods are supported
type mockS3 struct {
	NopAPI
	isErr    bool
	numCalls int
}

// IsErr returns whether the mock is configured to always fail
func (m mockS3) IsErr() bool {
	return m.isErr
}

// NumCalls returns how many times the method has been called
func (m mockS3) NumCalls() int {
	return m.numCalls
}

// SpyGetObject spies S3 GetObject and provides a specifies payload
type SpyGetObject struct {
	mockS3
	Payload []byte
}

// SpyGetObjectWithPayload will always provide the specified payload
func SpyGetObjectWithPayload(payload []byte) *SpyGetObject {
	return &SpyGetObject{Payload: payload}
}

// SpyGetObjectWithError which always fails
func SpyGetObjectWithError() *SpyGetObject {
	return &SpyGetObject{mockS3: mockS3{isErr: true}}
}

// GetObjectWithContext returns a reader from where to get the Payload
func (s *SpyGetObject) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, options ...request.Option) (*s3.GetObjectOutput, error) {
	s.numCalls++
	if s.IsErr() {
		return nil, errors.New("SpyGetObject error")
	}
	err := s.checkContext(ctx)
	if err != nil {
		return nil, err
	}
	out := &s3.GetObjectOutput{Body: ioutil.NopCloser(bytes.NewBuffer(s.Payload))}
	return out, nil
}

// SpyPutObject spies S3 PutObject recording the stored payload
type SpyPutObject struct {
	mockS3
	Payload []byte
}

// SpyPutObjectWithError returns a SpyPutObject which always fails
func SpyPutObjectWithError() *SpyPutObject {
	return &SpyPutObject{mockS3: mockS3{isErr: true}}
}

// PutObjectWithContext records the payload in the input
func (s *SpyPutObject) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, options ...request.Option) (*s3.PutObjectOutput, error) {
	s.numCalls++
	if s.IsErr() {
		return nil, errors.New("SpyPutObject error")
	}

	err := s.NopAPI.checkContext(ctx)
	if err != nil {
		return nil, err
	}
	s.Payload, _ = ioutil.ReadAll(input.Body)
	return s.mockS3.PutObjectWithContext(ctx, input)
}

// SpyDeleteObject spies S3 DeleteObject
type SpyDeleteObject struct {
	mockS3
}

// SpyDeleteObjectWithError returns a SpyDeleteObject which always fails
func SpyDeleteObjectWithError() *SpyDeleteObject {
	return &SpyDeleteObject{mockS3: mockS3{isErr: true}}
}

// DeleteObjectWithContext spies calls
func (s *SpyDeleteObject) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, options ...request.Option) (*s3.DeleteObjectOutput, error) {
	s.numCalls++
	if s.IsErr() {
		return nil, errors.New("SpyDeleteObject error")
	}
	return s.mockS3.DeleteObjectWithContext(ctx, input)
}
