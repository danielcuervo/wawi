package s3x

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// NopAPI mocks all the WithContext methods whose output is usually discarded (typically "writes")
// The methods fail if the context's Done() is closed
type NopAPI struct {
	s3iface.S3API
}

// PutObjectWithContext fails iff context is canceled
func (s *NopAPI) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, options ...request.Option) (*s3.PutObjectOutput, error) {
	err := s.checkContext(ctx)
	if err != nil {
		return nil, err
	}

	return &s3.PutObjectOutput{}, nil
}

// DeleteObjectWithContext iff context is canceled
func (s *NopAPI) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, options ...request.Option) (*s3.DeleteObjectOutput, error) {
	err := s.checkContext(ctx)
	if err != nil {
		return nil, err
	}
	return &s3.DeleteObjectOutput{}, nil
}

func (s *NopAPI) checkContext(ctx aws.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
