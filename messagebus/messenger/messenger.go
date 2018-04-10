package messenger

import "context"

type Message interface {
	Topic() string
	Payload() map[string]interface{}
}

type Logger interface {
	Log(msg Message)
}

type Handler interface {
	Handle(msg Message)
	Name() string
}

type Driver interface {
	Consume(topic string, ctx context.Context) error
	Dispatch(msg Message) error
	Receive() <-chan Message
}
