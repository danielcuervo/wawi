package messenger

import (
	"context"
)

// Driver of the tool you use for dispatching and consuming messages
type Driver interface {
	Consume(topic string, serviceID string, ctx context.Context) error
	Dispatch(msg Message) error
	Receive() <-chan Message
}
