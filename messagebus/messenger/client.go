package messenger

import (
	"context"
)

type client struct {
	driver    Driver
	logger    Logger
	consumers map[string]map[string]context.CancelFunc
}

func NewClient(driver Driver, logger Logger) *client {
	return &client{
		driver:    driver,
		logger:    logger,
		consumers: make(map[string]map[string]context.CancelFunc),
	}
}

func (c *client) Consume(topic string, handler Handler) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go c.driver.Consume(topic, ctx)

	go func() {
		defer cancelFunc()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.driver.Receive():
				handler.Handle(msg)
			}
		}
	}()

	if c.consumers[topic] == nil {
		c.consumers[topic] = make(map[string]context.CancelFunc)
	}

	c.consumers[topic][handler.Name()] = cancelFunc

	<-ctx.Done()
}

func (c *client) Dispatch(msg Message) {
	c.driver.Dispatch(msg)
}
