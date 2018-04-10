package messenger

import (
	"context"
)

type consumer struct {
	driver           ConsumerDriver
	handler          Handler
	receivedMessages chan Message
}

type ConsumerDriver interface {
	Listen(ctx context.Context)
	Receive() <-chan Message
}

func NewConsumer(driver ConsumerDriver, handler Handler) *consumer {
	return &consumer{
		driver:           driver,
		handler:          handler,
		receivedMessages: make(chan Message),
	}
}

func (c *consumer) Consume() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go c.driver.Listen(ctx)

	go func() {
		defer cancelFunc()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.driver.Receive():
				c.handler.Handle(msg)
			}
		}
	}()
}
