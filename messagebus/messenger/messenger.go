package messenger

import "context"

// Messages contain domain information and should be immutable objects
type Message interface {
	Topic() string
	Payload() map[string]interface{}
}

// How you want to log the messages going through this service
type Logger interface {
	Log(msg Message)
}

// Handler handle messages and should be immutable objects
type Handler interface {
	Handle(msg Message)
	Name() string
}

// Driver of the tool you use for dispatching and consuming messages
type Driver interface {
	Consume(topic string, ctx context.Context) error
	Dispatch(msg Message) error
	Receive() <-chan Message
}
