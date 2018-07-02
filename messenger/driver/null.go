package driver

import "github.com/danielcuervo/wawi/messenger"

type nullLogger struct {
}

// creates a pointer to a new instance of a null logger
func NewNullLogger() *nullLogger {
	return &nullLogger{}
}

func (nl *nullLogger) Log(message messenger.Message) {
}
