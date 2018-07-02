package driver

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/danielcuervo/wawi/messenger"
)

type kafkaDriver struct {
	address     string
	consumer    sarama.PartitionConsumer
	producer    sarama.AsyncProducer
	receivedMsg chan messenger.Message
}

// Creates a driver that consumes kafka messages
func NewKafkaDriver(address string) (*kafkaDriver, error) {
	producer, err := sarama.NewAsyncProducer([]string{address}, sarama.NewConfig())
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &kafkaDriver{address: address, producer: producer, receivedMsg: make(chan messenger.Message)}, nil
}

func (kd *kafkaDriver) Receive() <-chan messenger.Message {
	return kd.receivedMsg
}

func (kd *kafkaDriver) Consume(topic string, serviceID string, ctx context.Context) error {
	consumer, err := cluster.NewConsumer([]string{kd.address}, serviceID, []string{topic}, cluster.NewConfig())
	if err != nil {
		log.Println(err)
		return err
	}

	for {
		select {
		case err := <-consumer.Errors():
			log.Println(err.Error())
		case <-ctx.Done():
			return nil
		case msg := <-consumer.Messages():
			payload := &map[string]interface{}{}
			json.Unmarshal(msg.Value, payload)
			kd.receivedMsg <- &message{
				topic:   msg.Topic,
				payload: *payload,
			}
		}
	}

	return nil
}

func (kd *kafkaDriver) Dispatch(msg messenger.Message) error {
	kd.producer.Input() <- &sarama.ProducerMessage{
		Topic: msg.Topic(),
		Value: &payloadEncoder{msg.Payload()},
	}

	return nil
}

type payloadEncoder struct {
	Payload map[string]interface{}
}

func (pe *payloadEncoder) Length() int {
	encoded, err := json.Marshal(pe.Payload)
	if err != nil {
		return 0
	}

	return len(encoded)
}

func (pe *payloadEncoder) Encode() ([]byte, error) {
	encoded, err := json.Marshal(pe.Payload)
	return encoded, err
}

type message struct {
	topic   string
	payload map[string]interface{}
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Payload() map[string]interface{} {
	return m.payload
}
