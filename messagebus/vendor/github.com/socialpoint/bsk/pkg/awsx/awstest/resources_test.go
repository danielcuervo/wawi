package awstest_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/uuid"
	"github.com/stretchr/testify/assert"
	//"github.com/aws/aws-sdk-go/service/kinesis"
)

var resourceTypes = []string{
	s3.ServiceName,
	sqs.ServiceName,
	kms.ServiceName,
	dynamodb.ServiceName,
	awstest.SQSFifoServiceName,
	//kinesis.ServiceName,
}

func TestCreateResource(t *testing.T) {
	for _, res := range resourceTypes {
		awstest.AssertResourceExists(t, awstest.CreateResource(t, res), res)
	}
}

func TestKMSAliasCreatedForResource(t *testing.T) {
	t.Skip("There is an AWS limit, reconsider this test")
	assert := assert.New(t)

	keyID := awstest.CreateResource(t, kms.ServiceName)

	svc := kms.New(awstest.NewSession())
	res, err := svc.ListAliases(&kms.ListAliasesInput{Limit: aws.Int64(100)})

	assert.NoError(err)
	exists := false
	for _, a := range res.Aliases {
		if a.TargetKeyId != nil && *a.TargetKeyId == *keyID {
			exists = true
			break
		}
	}
	assert.True(exists)
}

func TestAssertResourceExists(t *testing.T) {
	mt := new(testing.T)

	for _, res := range resourceTypes {
		exists := awstest.AssertResourceExists(mt, aws.String(uuid.New()), res)
		assert.False(t, exists, res)
	}
}
