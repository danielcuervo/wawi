package messenger

import (
	"context"
	"sync"
)

// Messages contain domain information and should be immutable objects
type Message interface {
	Topic() string
	Payload() map[string]interface{}
}

// Handler handle messages and should be immutable objects
type Handler interface {
	Handle(msg Message)
	Name() string
}

type messenger struct {
	driver    Driver
	logger    Logger
	mutex     sync.Mutex
	consuming bool
	consumers map[string]map[string]consumer
}

type consumer struct {
	ServiceID  string
	Topic      string
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
		consumers: make(map[string]map[string]consumer),
	}
}

// Registers a consumer and starts consuming using the messenger driver
func (c *messenger) Consume(topic string, serviceID string, handler Handler, ctx context.Context) {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	c.mutex.Lock()
	if c.consumers[topic] == nil {
		c.consumers[topic] = make(map[string]consumer)
	}

	c.consumers[topic][handler.Name()] = consumer{
		Topic:      topic,
		ServiceID:  serviceID,
		Handler:    handler,
		CancelFunc: cancelFunc,
	}
	c.mutex.Unlock()

	c.driver.Consume(topic, serviceID, ctx)
}

// This starts loop to hook consumed messages through drivers and routing them to handlers, ideally this would
// be called before starting consuming
func (c *messenger) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		case msg := <-c.driver.Receive():
			topic := msg.Topic()
			c.logger.Log(msg, ctx)
			c.mutex.Lock()
			consumers := c.consumers[topic]
			c.mutex.Unlock()
			for _, consumer := range consumers {
				consumer.Handler.Handle(msg)
			}

		}
	}
}

// Dispatches a message
func (c *messenger) Dispatch(msg Message) {
	c.driver.Dispatch(msg)
}

// ServiceID function to give control over consumer stops
func (c *messenger) StopConsumer(topic string, handlerName string) {
	c.mutex.Lock()
	if c.consumers[topic] == nil {
		return
	}

	c.consumers[topic][handlerName].CancelFunc()
	c.consumers[topic] = make(map[string]consumer)
	c.mutex.Unlock()
}
