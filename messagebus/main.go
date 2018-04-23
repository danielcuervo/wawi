package main

import (
	"log"
	"strconv"

	"net"
	"time"

	"github.com/danielcuervo/wawi/messagebus/driver"
	"github.com/danielcuervo/wawi/messagebus/messenger"
)

func main() {
	//	messengerServer := messenger.NewHttpServer()
	//	messengerServer.Serve()
	ensureServicesAreAlive()

	stop := make(chan bool)
	kafkaDriver, err := driver.NewKafkaDriver("kafka:9092")
	if err != nil {
		log.Println(err.Error())
		return
	}

	client := messenger.NewMessenger(kafkaDriver, driver.NewNullLogger())
	go client.Consume("hello_world", &helloWorldHandler{})

	log.Println("Starting consumer")

	log.Println("Dispatching")
	for i := 0; i < 10; i++ {
		log.Println("Message" + strconv.Itoa(i))
		client.Dispatch(
			&helloWorld{
				Number: strconv.Itoa(i),
			},
		)
	}

	<-stop
}

func ensureServicesAreAlive() {
	for {
		conn, err := net.DialTimeout("tcp", "kafka:9092", time.Second)
		if conn != nil {
			return
		}

		log.Println(err.Error())
		time.Sleep(time.Second * 10)
	}
}

type helloWorld struct {
	Number string
}

func (m *helloWorld) Topic() string {
	return "hello_world"
}

func (hw *helloWorld) Payload() map[string]interface{} {
	return map[string]interface{}{
		"number": hw.Number,
	}
}

type helloWorldHandler struct {
}

func (hwh *helloWorldHandler) Handle(msg messenger.Message) {
	log.Println(msg.Topic())
}

func (hwh *helloWorldHandler) Name() string {
	return "hello_world_handler"
}
