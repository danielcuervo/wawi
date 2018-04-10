package main

import (
	"log"
	"strconv"

	"net"
	"time"

	"github.com/danielcuervo/wawi/messagebus/messenger"
	"github.com/socialpoint/bsk/pkg/uuid"
)

func main() {
	//	messengerServer := messenger.NewHttpServer()
	//	messengerServer.Serve()
	ensureServicesAreAlive()

	stop := make(chan bool)
	kafkaDriver, err := messenger.NewKafkaDriver(
		"kafka:9092",
		"hello_world",
	)
	if err != nil {
		log.Println(err.Error())
		return
	}
	consumer := messenger.NewConsumer(
		kafkaDriver,
		&helloWorldHandler{
			count: 0,
			stop:  stop,
		},
	)

	log.Println("Starting consumer")
	go consumer.Consume()

	log.Println("Dispatching")
	dispatcher := messenger.NewDispatcher(kafkaDriver)
	for i := 0; i < 10; i++ {
		log.Println("Message" + strconv.Itoa(i))
		dispatcher.Dispatch(
			&helloWorld{
				Uuid: uuid.New(),
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
	Uuid string
}

func (m *helloWorld) Topic() string {
	return "hello_world"
}

func (hw *helloWorld) Payload() map[string]interface{} {
	return map[string]interface{}{
		"uuid": hw.Uuid,
	}
}

type helloWorldHandler struct {
	count int
	stop  chan bool
}

func (hwh *helloWorldHandler) Handle(msg messenger.Message) {
	hwh.count++
	if hwh.count == 10 {
		log.Println("yes!")
		hwh.stop <- true
	}
}
