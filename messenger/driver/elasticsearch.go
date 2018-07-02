package driver

import (
	"context"

	"github.com/danielcuervo/wawi/messenger"
	"github.com/olivere/elastic"
)

type elasticSearchLogger struct {
	index     string
	indexType string
	client    *elastic.Client
}

// Creates a pointer to a new instance of a logger logging to elastic search
func NewElasticSearchLogger(host string, index string, indexType string) *elasticSearchLogger {
	client, _ := elastic.NewClient(
		elastic.SetURL(host),
	)

	return &elasticSearchLogger{
		index:     index,
		indexType: indexType,
		client:    client,
	}
}

func (el *elasticSearchLogger) Log(message messenger.Message, ctx context.Context) {
	log := elasticSearchMessage{Topic: message.Topic(), Payload: message.Payload()}

	el.client.Index().
		Index(el.index).
		Type(el.indexType).
		BodyJson(log).
		Do(ctx)
}

type elasticSearchMessage struct {
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
	Success bool                   `json:"success"`
}
