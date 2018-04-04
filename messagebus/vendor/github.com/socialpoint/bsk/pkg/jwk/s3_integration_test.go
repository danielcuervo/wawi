// +build integration

package jwk_test

import (
	"testing"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestS3ReaderAndWriter(t *testing.T) {
	assert := assert.New(t)

	bucket, cmk := setup(t)

	key, err := jwk.NewTestKey()
	assert.NoError(err)

	b := jwk.NewS3Bucket(awstest.NewSession(), bucket, uuid.New(), cmk, 5)

	// ops and assertions
	err = b.Write(key, key, key)
	assert.NoError(err)

	keys, err := b.Read()
	assert.NoError(err)

	assert.Len(keys, 3)
	//fix because created time has nil location but the one read from s3 chas UTC
	setUTCWhenEmpty(&key.NotBefore)
	setUTCWhenEmpty(&key.NotAfter)
	assert.EqualValues(key, keys[0])
}

func setup(t *testing.T) (bucket string, cmk string) {
	assert := assert.New(t)

	bucket = aws.StringValue(awstest.CreateResource(t, s3.ServiceName))
	_ = awstest.AssertResourceExists(t, aws.String(bucket), s3.ServiceName)

	cmk = aws.StringValue(awstest.CreateResource(t, kms.ServiceName))
	_ = awstest.AssertResourceExists(t, aws.String(cmk), kms.ServiceName)

	sess := awstest.NewSession()

	// Enable bucket versioning
	client := s3.New(sess)
	_, err := client.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	})
	assert.NoError(err)

	return bucket, cmk
}

// Though empty time is supposed to have UTC time (https://stackoverflow.com/questions/23051973/what-is-the-zero-value-for-time-time-in-go)
// in go 1.7 it internally contains a nil location
func setUTCWhenEmpty(t *time.Time) {
	//however we have to compare to UTC and not nil because Location() returns UTC when it's nil
	if t.IsZero() && t.Location() == time.UTC {
		*t = t.In(time.UTC)
	}
}
