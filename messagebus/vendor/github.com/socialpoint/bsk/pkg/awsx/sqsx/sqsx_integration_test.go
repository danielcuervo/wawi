// +build integration

package sqsx_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/awsx/sqsx"
	"github.com/stretchr/testify/assert"
)

func Test_Send_And_Receive(t *testing.T) {
	assert := assert.New(t)
	url := setupQueue(t, assert)
	sess := awstest.NewSession()

	payload := "random-payload:life-is-to-short-to-generate-a-really-random-payload-when-it-should-not-be-random-at-all"

	_, err := sqsx.SendMessage(context.Background(), sess, url, payload)
	assert.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), awstest.NewSession(), url, 2, 2)
	assert.NoError(err)

	assert.NotNil(msg)
	assert.Equal(payload, *msg.Body)
}

func Test_Delete(t *testing.T) {
	assert := assert.New(t)
	url := setupQueue(t, assert)
	sess := awstest.NewSession()

	payload := "random-payload:life-is-to-short-to-generate-a-really-random-payload-when-it-should-not-be-random-at-all"

	_, err := sqsx.SendMessage(context.Background(), sess, url, payload)
	assert.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), awstest.NewSession(), url, 2, 2)
	assert.NoError(err)

	assert.NotNil(msg)
	assert.Equal(payload, *msg.Body)

	err = sqsx.DeleteMessage(msg.ReceiptHandle, sess, url)
	assert.NoError(err)
}

func Test_ChangeMsgVisibilityTimeout(t *testing.T) {
	assert := assert.New(t)
	url := setupQueue(t, assert)
	sess := awstest.NewSession()

	payload := "random-payload:life-is-to-short-to-generate-a-really-random-payload-when-it-should-not-be-random-at-all"

	_, err := sqsx.SendMessage(context.Background(), sess, url, payload)
	assert.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), awstest.NewSession(), url, 300, 2)
	assert.NoError(err)
	assert.NotNil(msg)
	assert.Equal(payload, *msg.Body)

	_, err = sqsx.ChangeMsgVisibilityTimeout(context.Background(), awstest.NewSession(), url, msg.ReceiptHandle, 0)
	assert.NoError(err)
}

func setupQueue(t *testing.T, assert *assert.Assertions) (url string) {
	queue := awstest.CreateResource(t, sqs.ServiceName)
	awstest.AssertResourceExists(t, queue, sqs.ServiceName)

	url, err := sqsx.GetQueueURL(context.Background(), awstest.NewSession(), queue)
	assert.NoError(err)

	return
}

func Test_Send_And_Receive_From_FIFO(t *testing.T) {
	assert := assert.New(t)
	url := setupFIFOQueue(t, assert)
	sess := awstest.NewSessionWithRegion("us-east-2")

	payloadA := "payloadA"
	payloadB := "payloadB"
	group := "group"
	id1 := "1"
	id2 := "2"

	_, err := sqsx.SendFIFOMessage(context.Background(), sess, url, payloadA, group, id1)
	assert.NoError(err)

	_, err = sqsx.SendFIFOMessage(context.Background(), sess, url, payloadB, group, id2)
	assert.NoError(err)

	// Publish with the same deduplicationID -> message is not stored in the queue
	_, err = sqsx.SendFIFOMessage(context.Background(), sess, url, payloadB, group, id2)
	assert.NoError(err)

	msg, err := sqsx.ReceiveMessage(context.Background(), sess, url, 2, 2)
	assert.NoError(err)
	assert.NotNil(msg)

	if msg != nil {
		assert.Equal(payloadA, *msg.Body)

		err = sqsx.DeleteMessage(msg.ReceiptHandle, sess, url)
		assert.NoError(err)
	}

	msg, err = sqsx.ReceiveMessage(context.Background(), sess, url, 2, 2)
	assert.NoError(err)
	assert.NotNil(msg)

	if msg != nil {
		assert.Equal(payloadB, *msg.Body)

		err = sqsx.DeleteMessage(msg.ReceiptHandle, sess, url)
		assert.NoError(err)
	}

	// Duplicated message is not found
	msg, err = sqsx.ReceiveMessage(context.Background(), sess, url, 2, 2)
	assert.NoError(err)
	assert.Nil(msg)
}

func setupFIFOQueue(t *testing.T, assert *assert.Assertions) string {
	queue := awstest.CreateResource(t, awstest.SQSFifoServiceName)
	awstest.AssertResourceExists(t, queue, awstest.SQSFifoServiceName)

	url, err := sqsx.GetQueueURL(context.Background(), awstest.NewSessionWithRegion("us-east-2"), queue)
	assert.NoError(err)

	return url
}
