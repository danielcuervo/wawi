package driver

import "github.com/danielcuervo/wawi/messagebus/messenger"

type nullLogger struct {
}

func NewNullLogger() *nullLogger {
	return &nullLogger{}
}

func (nl *nullLogger) Log(message messenger.Message) {
}
