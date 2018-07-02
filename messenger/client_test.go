package messenger_test

import (
	"testing"

	"context"

	"sync"

	"github.com/Pallinder/go-randomdata"
	"github.com/danielcuervo/wawi/messenger"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	client := messenger.NewMessenger(&testDriver{}, &testLogger{})
	assert.NotEmpty(client)
}

func TestClientConsumeCallsDriver(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{Topics: make([]string, 0), ReceivedMessage: make(chan messenger.Message)}
	logger := &testLogger{}
	subjectUnderTest := messenger.NewMessenger(driver, logger)
	handler := &testHandler{}
	topic := "test"
	serviceID := "serviceA"
	ctx := context.Background()
	go subjectUnderTest.Consume(topic, serviceID, handler, ctx)
	<-driver.Receive()
	assert.Equal(1, driver.ConsumeCalls)
	assert.Equal(topic, driver.Topics[0])
}

func TestClientConsumeConcurrency(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{Topics: make([]string, 0), ReceivedMessage: make(chan messenger.Message)}
	logger := &testLogger{}
	subjectUnderTest := messenger.NewMessenger(driver, logger)
	handler := &testHandler{}
	topics := make(map[string]string)
	ctx := context.Background()
	serviceID := "service"
	for i := 0; i < 100; i++ {
		newTopic := randomdata.RandStringRunes(20)
		topics[newTopic] = newTopic
		go subjectUnderTest.Consume(newTopic, serviceID, handler, ctx)
	}

	for i := 0; i < 100; i++ {
		<-driver.Receive()
	}

	for _, topic := range topics {
		assert.Contains(driver.Topics, topic)
	}
	assert.Equal(100, driver.ConsumeCalls)
}

func TestClientStartCallsHandlerWhenConsumingAMessage(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{Topics: make([]string, 0), ReceivedMessage: make(chan messenger.Message)}
	logger := &testLogger{}
	subjectUnderTest := messenger.NewMessenger(driver, logger)
	handler := &testHandler{
		MessagesHandled: make(chan messenger.Message),
	}

	ctx := context.Background()

	topic := "test"
	serviceID := "serviceA"
	go subjectUnderTest.Start(ctx)
	go subjectUnderTest.Consume(topic, serviceID, handler, ctx)
	<-handler.MessagesHandled

	assert.Equal(1, driver.ConsumeCalls)
}

func TestClient_Start_It_Calls_Handler(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{Topics: make([]string, 0), ReceivedMessage: make(chan messenger.Message)}
	logger := &testLogger{}
	subjectUnderTest := messenger.NewMessenger(driver, logger)
	handler := &testHandler{
		MessagesHandled: make(chan messenger.Message),
	}

	ctx := context.Background()

	topic := "test"
	serviceID := "serviceA"
	go subjectUnderTest.Start(ctx)
	go subjectUnderTest.Consume(topic, serviceID, handler, ctx)
	<-handler.MessagesHandled

	assert.Equal(1, driver.ConsumeCalls)
}

type testDriver struct {
	mutex           sync.Mutex
	ConsumeCalls    int
	Topics          []string
	ReceivedMessage chan messenger.Message
}

func (td *testDriver) Consume(topic string, serviceID string, ctx context.Context) error {
	td.mutex.Lock()
	td.Topics = append(td.Topics, topic)
	td.ConsumeCalls++
	td.mutex.Unlock()
	td.ReceivedMessage <- &testMessage{}

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
	called int
}

func (t *testLogger) Log(msg messenger.Message, ctx context.Context) {
	t.called++
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
}

func (tm testMessage) Topic() string {
	return "test"
}

func (tm testMessage) Payload() map[string]interface{} {
	return map[string]interface{}{"test": "test"}
}
