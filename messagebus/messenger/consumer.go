package messenger

import (
	"context"
	"encoding/json"
	"log"

	"net"
	"time"

	"github.com/Shopify/sarama"
)

type kafkaConsumer struct {
}

func NewKafkaConsumer() *kafkaConsumer {
	return &kafkaConsumer{}
}

func (kc *kafkaConsumer) Consume(topic string, stop chan bool) {
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

	master, err := sarama.NewConsumer([]string{"kafka:9092"}, sarama.NewConfig())
	if err != nil {
		log.Print(err.Error())
		return
	}
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Print(err.Error())
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		i := 0
		for {
			select {
			case err := <-consumer.Errors():
				log.Println(err.Error())
			case <-ctx.Done():
				return
			case msg := <-consumer.Messages():
				type Message struct {
					Uuid string `json:"uuid"`
				}
				payload := &Message{}
				_ = json.Unmarshal(msg.Value, payload)
				i++
				if i == 10 {
					cancelFunc()
				}

			}
		}
	}()

	stop <- true
}
