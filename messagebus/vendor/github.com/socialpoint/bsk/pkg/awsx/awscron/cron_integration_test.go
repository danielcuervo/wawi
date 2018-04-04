// +build integration

package awscron_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint/bsk/pkg/awsx/awscron"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/awsx/sqsx"
	"github.com/stretchr/testify/assert"
)

func TestCronRunner(t *testing.T) {
	assert := assert.New(t)
	session := awstest.NewSession()
	now := time.Now()
	resources := []string{
		"arn:id-of-the-queue:queu-name",
	}

	url := getTestSchedulerQueue(t)
	event, err := awscron.SendEvent(context.Background(), session, url, now, resources)
	assert.NoError(err)

	events := make(chan *awscron.Event)

	f := func(ev *awscron.Event) error {
		events <- ev
		return nil
	}

	e := func(err error) {}

	runner := awscron.CronRunner(awstest.NewSession(), url, f, e)

	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)

	received := <-events

	assert.Equal(event.Time.Unix(), received.Time.Unix())
	assert.Equal(event.Resources, received.Resources)

	cancel()
}

func getTestSchedulerQueue(t *testing.T) string {
	assert := assert.New(t)

	queue := awstest.CreateResource(t, sqs.ServiceName)
	awstest.AssertResourceExists(t, queue, sqs.ServiceName)

	url, err := sqsx.GetQueueURL(context.Background(), awstest.NewSession(), queue)
	assert.NoError(err)

	return url
}
