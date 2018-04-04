package main

import (
	"log"
	"strconv"

	"github.com/danielcuervo/wawi/messagebus/messenger"
	"github.com/socialpoint/bsk/pkg/uuid"
)

func main() {
	//	messengerServer := messenger.NewHttpServer()
	//	messengerServer.Serve()
	consumer := messenger.NewKafkaConsumer()
	stop := make(chan bool)
	log.Println("Starting consumer")
	go consumer.Consume("hello_world", stop)

	log.Println("Dispatching")
	dispatcher := messenger.NewKafkaDispatcher()
	for i := 0; i < 10; i++ {
		log.Println("Message" + strconv.Itoa(i))
		dispatcher.Dispatch(
			&messenger.HelloWorld{
				Uuid: uuid.New(),
			},
		)
	}

	for {
		if <-stop {
			break
		}
	}
}
