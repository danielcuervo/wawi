// +build integration

package s3x_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/awsx/s3x"
)

func TestDownload(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	const payload = "u42\nu43"
	const key = "push-notifications/pro/dc/uuid-9998"

	sess := awstest.NewSession()
	bucket := *awstest.CreateResource(t, s3.ServiceName)
	awstest.AssertResourceExists(t, &bucket, s3.ServiceName)

	writeToS3(t, sess, bucket, key, payload)
	defer deleteS3File(t, sess, bucket, key)

	s3Cli := s3.New(sess)
	s3xAPI := s3x.New(s3Cli)
	usersReader, err := s3xAPI.Download(context.Background(), s3x.NewURI(bucket, key))
	a.NoError(err)

	readUsers, err := ioutil.ReadAll(usersReader)
	a.NoError(err)
	a.Equal(payload, string(readUsers))
	a.NoError(usersReader.Close())
}

func TestUpload(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	const payload = "u42\nu41"
	const key = "push-notifications/pro/dc/uuid-9999"

	sess := awstest.NewSession()
	bucket := *awstest.CreateResource(t, s3.ServiceName)
	awstest.AssertResourceExists(t, &bucket, s3.ServiceName)
	s3Cli := s3.New(sess)
	s3xAPI := s3x.New(s3Cli)

	err := s3xAPI.Upload(context.Background(), s3x.NewURI(bucket, key), strings.NewReader(payload))
	a.NoError(err)

	defer deleteS3File(t, sess, bucket, key)

	a.Equal(payload, readFromS3(t, sess, bucket, key))
}

func TestDelete_fileNotAvailable(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	sess := awstest.NewSession()
	bucket := *awstest.CreateResource(t, s3.ServiceName)
	awstest.AssertResourceExists(t, &bucket, s3.ServiceName)
	s3Cli := s3.New(sess)
	s3xAPI := s3x.New(s3Cli)

	rand.Seed(time.Now().UnixNano())
	rndKey := fmt.Sprintf("push-notifications/pro/dc/TestDelete%d", rand.Int())
	uri := s3x.NewURI(bucket, rndKey)

	//succeeds when it does not exist
	err := s3xAPI.Delete(context.Background(), uri)
	a.NoError(err)

	//succeeds when already deleted
	err = s3xAPI.Delete(context.Background(), uri)
	a.NoError(err)
}

func TestDelete_fileAvailable(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	sess := awstest.NewSession()
	bucket := *awstest.CreateResource(t, s3.ServiceName)
	awstest.AssertResourceExists(t, &bucket, s3.ServiceName)
	s3Cli := s3.New(sess)
	s3xAPI := s3x.New(s3Cli)

	rand.Seed(time.Now().UnixNano())
	//using fixed key to avoid paying for many deleted markers
	const fixedKey = "push-notifications/pro/dc/TestDelete"
	rndKey := fmt.Sprintf("%s%d", fixedKey, rand.Int())
	uri := s3x.NewURI(bucket, rndKey)

	//succeeds when it exists
	writeToS3(t, sess, bucket, fixedKey, "data")
	defer deleteS3File(t, sess, bucket, rndKey)

	err := s3xAPI.Delete(context.Background(), uri)
	a.NoError(err)

	//not accessible after deleted
	_, err = s3xAPI.Download(context.Background(), uri)
	a.Equal(s3.ErrCodeNoSuchKey, err.(awserr.Error).Code())
}

func TestErrorContainsContext(t *testing.T) {

	const bucket = "non_existing_bucket"
	const key = "push-notifications/pro/dc/uuid-9998"

	sess := awstest.NewSession()
	s3Cli := s3.New(sess)
	s3xAPI := s3x.New(s3Cli)

	tests := []struct {
		name string
		f    func(t *testing.T) error
	}{
		{
			name: "Download",
			f: func(t *testing.T) error {
				usersReader, err := s3xAPI.Download(context.Background(), s3x.NewURI("non_existing_bucket", "non_existing_key"))
				assert.Nil(t, usersReader)
				return err
			},
		},
		{
			name: "Upload",
			f: func(*testing.T) error {
				return s3xAPI.Upload(context.Background(), s3x.NewURI("non_existing_bucket", "non_existing_key"), nil)
			},
		},
		{
			name: "Delete",
			f: func(*testing.T) error {
				return s3xAPI.Delete(context.Background(), s3x.NewURI("non_existing_bucket", "non_existing_key"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			err := test.f(t)

			a.Error(err)
			a.Contains(err.Error(), "non_existing_bucket")
			a.Contains(err.Error(), "non_existing_key")
			a.Contains(err.Error(), test.name)
		})
	}
}

func readFromS3(t *testing.T, sess client.ConfigProvider, bucket string, s3key string) string {
	cli := s3.New(sess)
	output, err := cli.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(string(s3key)),
	})
	require.NoError(t, err)
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(output.Body)
	assert.NoError(t, err)
	return buf.String()
}

func deleteS3File(t *testing.T, sess client.ConfigProvider, bucket string, s3key string) {
	cli := s3.New(sess)
	_, err := cli.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(string(s3key)),
	})
	require.NoError(t, err)
}

func writeToS3(t *testing.T, sess client.ConfigProvider, bucket string, s3key string, data string) {
	cli := s3.New(sess)
	_, err := cli.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(string(s3key)),
		Body:   bytes.NewReader([]byte(data)),
	})
	require.NoError(t, err)
}
