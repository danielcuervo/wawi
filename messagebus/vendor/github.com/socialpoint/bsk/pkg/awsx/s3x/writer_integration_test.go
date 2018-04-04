// +build integration

package s3x_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/awsx/s3x"
	"github.com/socialpoint/bsk/pkg/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3writer_Write(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	sess := awstest.NewSession()

	bucket := awstest.CreateResource(t, s3.ServiceName)
	key := "writer-test-" + ulid.New()
	dst := s3x.S3Writer(context.Background(), s3manager.NewUploader(sess), *bucket, key)

	n := 10
	token := make([]byte, n)
	rand.Read(token)

	src := bytes.NewBuffer(token)

	written, err := io.Copy(dst, src)
	a.EqualValues(n, written)
	a.NoError(err)

	a.NoError(dst.Close())

	api := s3x.New(s3.New(sess))
	reader, err := api.Download(context.Background(), s3x.NewURI(*bucket, key))
	r.NoError(err)

	data, err := ioutil.ReadAll(reader)
	a.NoError(err)
	a.Equal(token, data)
	a.NoError(reader.Close())
}

func TestS3writer_Write_Failures_1part(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	sess := awstest.NewSession()

	bucket := "this-bucket-should-not-exists-" + ulid.New()
	key := "writer-test-" + ulid.New()
	dst := s3x.S3Writer(context.Background(), s3manager.NewUploader(sess), bucket, key)

	//for small payloads...
	n := int64(10)
	token := make([]byte, n)
	rand.Read(token)

	src := bytes.NewBuffer(token)
	written, err := io.Copy(dst, src)
	a.NoError(err)

	//... we don't get the error until closing the writer
	err = dst.Close()
	a.Equal(s3.ErrCodeNoSuchBucket, err.(awserr.Error).Code())
	a.EqualValues(10, written)
	assertDecoratedError(t, err, bucket, key)
}

func TestS3writer_Write_Failures_multipart(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	sess := awstest.NewSession()

	bucket := "this-bucket-should-not-exists-" + ulid.New()
	key := "writer-test-" + ulid.New()
	dst := s3x.S3Writer(context.Background(), s3manager.NewUploader(sess), bucket, key)

	//when multipart upload is required...
	const n = s3manager.MinUploadPartSize + 1
	token := make([]byte, n)
	rand.Read(token)

	//the writer client gets a meaningful error when writing
	src := bytes.NewBuffer(token)
	written, err := io.Copy(dst, src)
	a.Equal(s3.ErrCodeNoSuchBucket, err.(awserr.Error).Code())
	assertDecoratedError(t, err, bucket, key)

	a.NoError(dst.Close())
	a.EqualValues(s3manager.MinUploadPartSize, written)
}

func assertDecoratedError(t *testing.T, err error, bucket string, key string) {
	a := assert.New(t)
	a.Contains(err.Error(), s3x.NewURI(bucket, key).String())
	a.Contains(err.Error(), "s3writer")
}
