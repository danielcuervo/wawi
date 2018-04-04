package awscron

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint/bsk/pkg/awsx/sqsx"
	"github.com/socialpoint/bsk/pkg/server"
	"github.com/socialpoint/bsk/pkg/uuid"
)

// Event represents a CloudWatch scheduled event payload
// See https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/EventTypes.html#schedule_event_type
type Event struct {
	ID        string
	Source    string
	Account   string
	Time      time.Time
	Region    string
	Resources []string
}

// OnEvent is the type that callback functions must satisfy
type OnEvent func(*Event) error

// CronRunner returns a runner that watch a queue with CloudWatch scheduled events
func CronRunner(p client.ConfigProvider, url string, oe OnEvent, e sqsx.OnError) server.Runner {
	callback := func(msg *sqs.Message) error {
		event, err := sqsMessageToEvent(msg)
		if err != nil {
			return err
		}

		return oe(event)
	}

	return server.RunnerFunc(func(ctx context.Context) {
		sqsx.Watch(ctx, p, url, callback, e)
	})
}

// SendEvent delivers a message to the specified queue, simulating the triggering of an scheduled event
// at the given time
func SendEvent(ctx context.Context, p client.ConfigProvider, url string, t time.Time, resources []string) (*Event, error) {
	event := &Event{
		ID:        uuid.New(),
		Time:      t,
		Source:    "aws.events",
		Resources: resources,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	if _, err := sqsx.SendMessage(ctx, p, url, string(body)); err != nil {
		return nil, err
	}

	return event, nil
}

func sqsMessageToEvent(msg *sqs.Message) (*Event, error) {
	event := &Event{}
	err := json.Unmarshal([]byte(aws.StringValue(msg.Body)), event)

	return event, err
}
