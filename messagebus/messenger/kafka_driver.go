package messenger

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
)

type kafkaDriver struct {
	consumer    sarama.PartitionConsumer
	producer    sarama.AsyncProducer
	receivedMsg chan Message
}

// Creates a driver that consumes kafka messages
func NewKafkaDriver(address string, topic string) (*kafkaDriver, error) {
	master, err := sarama.NewConsumer([]string{address}, sarama.NewConfig())
	if err != nil {
		return nil, err
	}
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewAsyncProducer([]string{address}, sarama.NewConfig())
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &kafkaDriver{consumer: consumer, producer: producer, receivedMsg: make(chan Message)}, nil
}

func (kd *kafkaDriver) Receive() <-chan Message {
	return kd.receivedMsg
}

func (kd *kafkaDriver) Listen(ctx context.Context) {
	go func() {
		for {
			select {
			case err := <-kd.consumer.Errors():
				log.Println(err.Error())
			case <-ctx.Done():
				return
			case msg := <-kd.consumer.Messages():
				payload := &map[string]interface{}{}
				json.Unmarshal(msg.Value, payload)
				kd.receivedMsg <- &message{
					topic:   msg.Topic,
					payload: *payload,
				}
			}
		}
	}()
}

func (kd *kafkaDriver) Dispatch(msg Message) {
	kd.producer.Input() <- &sarama.ProducerMessage{
		Topic: msg.Topic(),
		Value: NewPayloadEncoder(msg.Payload()),
	}
}

type payloadEncoder struct {
	Payload map[string]interface{}
}

func NewPayloadEncoder(payload map[string]interface{}) *payloadEncoder {
	return &payloadEncoder{Payload: payload}
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
