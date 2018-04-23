package messenger

import (
	"context"

	"sync"
)

type messenger struct {
	driver    Driver
	logger    Logger
	mutex     sync.Mutex
	consuming bool
	registry  map[string]map[string]register
}

type register struct {
	Handler    Handler
	CancelFunc context.CancelFunc
}

// Client to use the messenger it abstracts consuming/dispatching from the tool of choice and holds ebs functionality
func NewMessenger(driver Driver, logger Logger) *messenger {
	return &messenger{
		driver:    driver,
		logger:    logger,
		mutex:     sync.Mutex{},
		consuming: false,
		registry:  make(map[string]map[string]register),
	}
}

// Creates a consumer to consume a topic and passes it to the handler
// stores it's cancel func in the consumers
func (c *messenger) Consume(topic string, handler Handler) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	c.register(topic, handler, cancelFunc)
	go c.driver.Consume(topic, ctx)
	c.start(ctx)
}

func (c *messenger) register(topic string, handler Handler, cancelFunc context.CancelFunc) {
	c.mutex.Lock()
	if c.registry[topic] == nil {
		c.registry[topic] = make(map[string]register)
	}

	c.registry[topic][handler.Name()] = register{
		Handler:    handler,
		CancelFunc: cancelFunc,
	}
	c.mutex.Unlock()
}

func (c *messenger) start(ctx context.Context) {
	c.mutex.Lock()
	if c.consuming {
		c.mutex.Unlock()
		return
	}

	c.consuming = true
	c.mutex.Unlock()

	for {
		select {
		case <-ctx.Done():
			break
		case msg := <-c.driver.Receive():
			topic := msg.Topic()
			c.mutex.Lock()
			handlers := c.registry[topic]
			c.mutex.Unlock()
			for _, handler := range handlers {
				handler.Handler.Handle(msg)
			}

		}
	}
}

// Dispatches a message
func (c *messenger) Dispatch(msg Message) {
	c.driver.Dispatch(msg)
}

// Service function to give control over consumer stops
func (c *messenger) StopConsumer(topic string, handlerName string) {
	c.mutex.Lock()
	if c.registry[topic] == nil {
		return
	}

	c.registry[topic][handlerName].CancelFunc()
	c.registry[topic] = make(map[string]register)
	c.mutex.Unlock()
}
