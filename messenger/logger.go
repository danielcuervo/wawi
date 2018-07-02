package messenger

import (
	"context"
)

// How you want to log the messages going through this client library
type Logger interface {
	Log(msg Message, ctx context.Context)
}
