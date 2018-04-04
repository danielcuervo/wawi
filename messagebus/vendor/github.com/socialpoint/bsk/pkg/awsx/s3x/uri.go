package s3x

import (
	"fmt"
	"net/url"
	"strings"
)

const uriScheme = "s3"

// URI is an light immutable class which manages S3Uri's
// See http://docs.aws.amazon.com/cli/latest/reference/s3/index.html#path-argument-type
// and http://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/services/s3/AmazonS3URI.html
type URI struct {
	bucket string
	key    string
}

// Bucket returns the s3 bucket name with no "/"
func (u URI) Bucket() string {
	return u.bucket
}

//Key contains object name, typically after its prefix path (using / as separator)
func (u URI) Key() string {
	return u.key
}

// NewURI creates a URI from a bucket and a key
func NewURI(bucket string, key string) *URI {
	return &URI{bucket: bucket, key: key}
}

// ConcatURI creates a new URI appending a subKey to an URI
func ConcatURI(u URI, subKey string) *URI {
	f := "%s/%s"
	if strings.HasSuffix(u.key, "/") {
		f = "%s%s"
	}
	return &URI{bucket: u.bucket, key: fmt.Sprintf(f, u.key, subKey)}
}

// ParseURI parses "s3://bucket/key"
// Key may be blank, but bucket must be followed by with /
func ParseURI(rawURI string) (*URI, error) {
	url, err := url.Parse(rawURI)
	if err != nil {
		return nil, fmt.Errorf("'%s' is not an URI ", rawURI)
	}
	if url.Scheme != uriScheme {
		return nil, fmt.Errorf("uri '%s' should have scheme %s", rawURI, uriScheme)
	}
	if url.Host == "" {
		return nil, fmt.Errorf("uri 's3://%s' does not have a bucket", rawURI)
	}
	if url.Path == "" {
		return nil, fmt.Errorf("uri 's3://%s' does not have a key", rawURI)
	}

	return &URI{
		bucket: url.Host,
		//skip initial /
		key: url.Path[1:len(url.Path)],
	}, nil
}

// String implements Stringer
func (u URI) String() string {
	return fmt.Sprintf("%s://%s/%s", uriScheme, u.bucket, u.key)
}
