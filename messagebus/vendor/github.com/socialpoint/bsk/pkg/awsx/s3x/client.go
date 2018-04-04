package s3x

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

//In the future we can perform validations, auditing, migrate s3iface.S3API mocks, implement "Exists(S3URI) bool"

// s3client implements s3x.API
type s3client struct {
	s3Api s3iface.S3API
}

// New creates an API for S3
func New(s3Api s3iface.S3API) API {
	return &s3client{s3Api: s3Api}
}

// Upload stores the contents reader in s3
// s3 needs input to implement Seeker to calculate the length, to sign the messaging
// Consider using s3 manager for large data (>5MB) or when Seeker is not feasible
func (c *s3client) Upload(ctx context.Context, uri *URI, input io.ReadSeeker) error {
	_, err := c.s3Api.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(uri.Bucket()),
		Key:    aws.String(uri.Key()),
		Body:   input,
	})
	return decorateError(err, "Upload", uri)
}

// Download returns a reader from where to get a file contents.
// Use s3manager downloader to efficiently download large files (at least 5MB)
func (c *s3client) Download(ctx context.Context, uri *URI) (io.ReadCloser, error) {

	out, err := c.s3Api.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(uri.Bucket()),
		Key:    aws.String(uri.Key()),
	})
	if err != nil {
		return nil, decorateError(err, "Download", uri)
	}
	return out.Body, nil
}

// Delete deletes a file succeeding even if it does not exist or is already marked as deleted
func (c *s3client) Delete(ctx context.Context, uri *URI) error {
	_, err := c.s3Api.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(uri.Bucket()),
		Key:    aws.String(uri.Key()),
	})
	return decorateError(err, "Delete", uri)
}
