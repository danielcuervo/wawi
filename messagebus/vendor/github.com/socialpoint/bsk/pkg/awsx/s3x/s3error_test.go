package s3x

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"
)

func TestDecorateError_awserr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	uri := NewURI("bucket", "key")
	orig := awserr.New("codeOrig", "msgOrg", nil)
	awsError := awserr.New("code1", "msg1", orig)
	decorated := decorateError(awsError, "writing", uri)

	decoratedAws, ok := decorated.(awserr.Error)
	a.True(ok)
	a.Contains(decoratedAws.Error(), "writing")
	a.Contains(decoratedAws.Error(), uri.String())
	a.Contains(decoratedAws.Error(), "msg1")
	a.Contains(decoratedAws.Error(), "msgOrg")

	a.Equal("code1", decoratedAws.Code())
	a.Equal(orig, decoratedAws.OrigErr())
}

func TestDecorateError_nonAwserr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	uri := NewURI("bucket", "key")
	nonAwsError := fmt.Errorf("msg1")

	decorated := decorateError(nonAwsError, "writing", uri)

	_, ok := decorated.(awserr.Error)
	a.False(ok)
	a.Equal(nonAwsError, decorated)
}
