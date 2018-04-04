package messenger

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/Shopify/sarama"
)

type Dispatcher interface {
	Dispatch(msg Message)
}

type kafkaDispatcher struct {
}

func NewKafkaDispatcher() *kafkaDispatcher {
	return &kafkaDispatcher{}
}

func (kd *kafkaDispatcher) Dispatch(msg Message) {
	for {
		conn, err := net.DialTimeout("tcp", "kafka:9092", time.Second)
		if conn != nil {
			conn.Close()
			break
		}

		if err != nil {
			log.Println(err.Error())
		}
		time.Sleep(time.Second * 5)
	}

	producer, err := sarama.NewAsyncProducer([]string{"kafka:9092"}, sarama.NewConfig())
	if err != nil {
		log.Println(err.Error())
		return
	}

	producer.Input() <- &sarama.ProducerMessage{
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
