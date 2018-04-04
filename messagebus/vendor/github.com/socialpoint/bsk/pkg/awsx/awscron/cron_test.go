package awscron_test

import (
	"testing"

	"encoding/json"

	"github.com/socialpoint/bsk/pkg/awsx/awscron"
	"github.com/stretchr/testify/assert"
)

func TestEventJSONUnmarshaling(t *testing.T) {
	assert := assert.New(t)

	data := `
	{
		"id": "53dc4d37-cffa-4f76-80c9-8b7d4a4d2eaa",
		"detail-type": "Scheduled Event",
		"source": "aws.events",
		"account": "123456789012",
		"time": "2015-10-08T16:53:06Z",
		"region": "us-east-1",
		"resources": [ "arn:aws:events:us-east-1:123456789012:rule/MyScheduledRule" ],
		"detail": {}
	}
	`

	var event awscron.Event

	err := json.Unmarshal([]byte(data), &event)
	assert.NoError(err)

	assert.Equal("53dc4d37-cffa-4f76-80c9-8b7d4a4d2eaa", event.ID)
	assert.Equal("aws.events", event.Source)
}
