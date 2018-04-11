package messenger

import (
	"context"
	"sync"
)

type client struct {
	driver        Driver
	logger        Logger
	registerMutex sync.Mutex
	consumers     map[string]map[string]context.CancelFunc
}

func NewClient(driver Driver, logger Logger) *client {
	return &client{
		driver:        driver,
		logger:        logger,
		registerMutex: sync.Mutex{},
		consumers:     make(map[string]map[string]context.CancelFunc),
	}
}

func (c *client) Consume(topic string, handler Handler) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go c.driver.Consume(topic, ctx)
	c.register(topic, handler, cancelFunc)

	for {
		select {
		case <-ctx.Done():
			break
		case msg := <-c.driver.Receive():
			handler.Handle(msg)
		}
	}
}

func (c *client) Dispatch(msg Message) {
	c.driver.Dispatch(msg)
}
func (c *client) register(topic string, handler Handler, cancelFunc context.CancelFunc) {
	c.registerMutex.Lock()
	defer c.registerMutex.Unlock()
	if c.consumers[topic] == nil {
		c.consumers[topic] = make(map[string]context.CancelFunc)
	}

	c.consumers[topic][handler.Name()] = cancelFunc
}
