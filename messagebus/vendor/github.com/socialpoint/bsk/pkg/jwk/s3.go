package jwk

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
)

// S3Reader returns a reader that reads keys from S3.
func S3Reader(p client.ConfigProvider, bucket, object string, maxReads int64) Reader {
	return ReaderFunc(func() ([]*Key, error) {
		client := s3crypto.NewDecryptionClient(p)

		res, err := client.S3Client.ListObjectVersions(&s3.ListObjectVersionsInput{
			Bucket:  aws.String(bucket),
			Prefix:  aws.String(object),
			MaxKeys: aws.Int64(maxReads),
		})
		if err != nil {
			return nil, err
		}

		keys := []*Key{}
		for _, v := range res.Versions {
			key, err := s3GetObject(client, bucket, object, *v.VersionId)
			if err != nil {
				return nil, err
			}

			keys = append(keys, key)
		}

		return keys, nil
	})
}

// S3Writer returns a writer that writes keys to S3, encrypted with the given customer master key.
func S3Writer(p client.ConfigProvider, bucket, object, cmkID string) Writer {
	return WriterFunc(func(keys ...*Key) error {
		handler := s3crypto.NewKMSKeyGenerator(kms.New(p), cmkID)
		client := s3crypto.NewEncryptionClient(p, s3crypto.AESGCMContentCipherBuilder(handler))

		for _, key := range keys {
			m, err := json.Marshal(key)
			if err != nil {
				return err
			}

			_, err = client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(object),
				Body:   bytes.NewReader(m),
			})
			if err != nil {
				return err
			}
		}

		return nil

	})
}

// NewS3Bucket returns a bucket that reads/writes from/to S3 in a secure way
func NewS3Bucket(sess *session.Session, bucket, key, cmkID string, maxReads int64) ReadWriter {
	return &s3bucket{
		bucket:   bucket,
		key:      key,
		cmkID:    cmkID,
		maxReads: maxReads,
		session:  sess,
	}
}

// s3bucket represents a bucket on S3 that holds Keys
type s3bucket struct {
	bucket, key string
	maxReads    int64
	cmkID       string
	session     *session.Session
}

func (b *s3bucket) Read() ([]*Key, error) {
	return S3Reader(b.session, b.bucket, b.key, b.maxReads).Read()
}

func (b *s3bucket) Write(keys ...*Key) error {
	return S3Writer(b.session, b.bucket, b.key, b.cmkID).Write(keys...)
}

func s3GetObject(client *s3crypto.DecryptionClient, bucket, object, versionID string) (*Key, error) {
	res, err := client.GetObject(&s3.GetObjectInput{
		Bucket:    aws.String(bucket),
		Key:       &object,
		VersionId: &versionID,
	})
	if err != nil {
		return nil, err
	}

	k := &Key{}
	err = json.NewDecoder(res.Body).Decode(k)
	if err != nil {
		return nil, err
	}

	return k, nil
}
