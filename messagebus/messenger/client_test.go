package messenger_test

import (
	"testing"

	"context"

	"sync"

	"github.com/Pallinder/go-randomdata"
	"github.com/danielcuervo/wawi/messagebus/messenger"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	client := messenger.NewMessenger(&testDriver{}, &testLogger{})
	assert.NotEmpty(client)
}

func TestClient_Consume_It_Calls_Handler(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{Topics: make([]string, 0), ReceivedMessage: make(chan messenger.Message)}
	client := messenger.NewMessenger(driver, &testLogger{})
	handler := &testHandler{
		MessagesHandled: make(chan messenger.Message),
	}
	topic := "test"
	go client.Consume(topic, handler)
	msg := <-handler.MessagesHandled
	assert.Equal(1, driver.ConsumeCalls)
	assert.Equal(topic, driver.Topics[0])
	assert.Equal(topic, msg.Topic())
	client.StopConsumer(topic, handler.Name())
}

func TestClient_Consume_Concurrency(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{Topics: make([]string, 0), ReceivedMessage: make(chan messenger.Message)}
	client := messenger.NewMessenger(driver, &testLogger{})
	handler := &testHandler{
		MessagesHandled: make(chan messenger.Message),
	}
	topics := make(map[string]string)
	for i := 0; i < 100; i++ {
		newTopic := randomdata.RandStringRunes(20)
		topics[newTopic] = newTopic
		go client.Consume(newTopic, handler)
	}
	counter := 0
	for counter < 100 {
		msg := <-handler.MessagesHandled
		counter++
		assert.Contains(topics, msg.Topic())
	}

	assert.Equal(100, driver.ConsumeCalls)
	for topic := range topics {
		assert.Contains(driver.GetTopics(), topic)
		client.StopConsumer(topic, handler.Name())
	}
}

type testDriver struct {
	mutex           sync.Mutex
	ConsumeCalls    int
	Topics          []string
	ReceivedMessage chan messenger.Message
}

func (td *testDriver) Consume(topic string, ctx context.Context) error {
	td.mutex.Lock()
	td.Topics = append(td.Topics, topic)
	td.ConsumeCalls++
	td.mutex.Unlock()
	td.ReceivedMessage <- testMessage{topic: topic}
	return nil
}

func (td *testDriver) Dispatch(msg messenger.Message) error {
	panic("implement me")
}

func (td *testDriver) Receive() <-chan messenger.Message {
	return td.ReceivedMessage
}

func (td *testDriver) GetTopics() []string {
	td.mutex.Lock()
	defer td.mutex.Unlock()
	return td.Topics
}

type testLogger struct {
}

func (t *testLogger) Log(msg messenger.Message) {
	panic("implement me")
}

type testHandler struct {
	mutex           sync.RWMutex
	HandleCalls     int
	MessagesHandled chan messenger.Message
}

func (th *testHandler) Handle(msg messenger.Message) {
	th.mutex.Lock()
	th.HandleCalls++
	th.mutex.Unlock()
	th.MessagesHandled <- msg
}

func (th *testHandler) Name() string {
	return "test_handler"
}

type testMessage struct {
	topic   string
	payload map[string]interface{}
}

func (tm testMessage) Topic() string {
	return tm.topic
}

func (tm testMessage) Payload() map[string]interface{} {
	return tm.payload
}
