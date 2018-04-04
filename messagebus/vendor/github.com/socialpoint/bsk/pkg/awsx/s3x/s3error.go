package s3x

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// Decorates an aws error adding context about what the application was doing (eg. writing) and object (eg. URI)
// If err is awserr.Error, it will have the same Code() and OrigErr() but Message() and Error() will contain the context
// Otherwise, it will return the argument error as is (we don't use errors.Wrapf to avoid hiding eg. io.EOF)
func decorateError(err error, doing string, object fmt.Stringer) error {
	switch e := err.(type) {
	case awserr.Error:
		msg := fmt.Sprintf("%s %s: %s", doing, object.String(), e.Message())
		return awserr.New(e.Code(), msg, e.OrigErr())
	default:
		return err
	}
}
