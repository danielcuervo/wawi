package messenger_test

import (
	"testing"

	"context"

	"github.com/danielcuervo/wawi/messagebus/messenger"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	client := messenger.NewClient(&testDriver{}, &testLogger{})
	assert.NotEmpty(client)
}

func TestClient_Consume_It_Calls_Handler(t *testing.T) {
	assert := assert.New(t)
	driver := &testDriver{}
	client := messenger.NewClient(driver, &testLogger{})
	handler := &testHandler{
		MessagesHandled: make(map[string]messenger.Message),
	}
	go client.Consume("test", handler)
	assert.Len(handler.MessagesHandled, 1)
	assert.Equal(1, driver.ConsumeCalls)
	assert.Equal("test", driver.Topics[0])

}

type testDriver struct {
	ConsumeCalls int
	Topics       map[int]string
}

func (td *testDriver) Consume(topic string, ctx context.Context) error {
	td.Topics[td.ConsumeCalls] = topic
	td.ConsumeCalls++
}

func (td *testDriver) Dispatch(msg messenger.Message) error {
	panic("implement me")
}

func (td *testDriver) Receive() <-chan messenger.Message {
	panic("implement me")
}

type testLogger struct {
}

func (t *testLogger) Log(msg messenger.Message) {
	panic("implement me")
}

type testHandler struct {
	MessagesHandled map[string]messenger.Message
}

func (th *testHandler) Handle(msg messenger.Message) {
	th.MessagesHandled[msg.Topic()] = msg
}

func (th *testHandler) Name() string {
	return "test_handler"
}
