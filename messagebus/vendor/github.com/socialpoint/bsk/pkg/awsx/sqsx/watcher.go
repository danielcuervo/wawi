package sqsx

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/socialpoint/bsk/pkg/server"
)

// OnMessage is the type that callback functions must satisfy
type OnMessage func(msg *sqs.Message) error

// OnError is the type that errors callback function must satisfy
type OnError func(error)

// WatchRunner returns a runner that watch a queue, using the runner's context
func WatchRunner(p client.ConfigProvider, url string, f OnMessage, e OnError) server.Runner {
	return server.RunnerFunc(func(ctx context.Context) {
		Watch(ctx, p, url, f, e)
	})
}

// Watch watches a queue a call the callback function on messages or errors
func Watch(ctx context.Context, p client.ConfigProvider, url string, f OnMessage, e OnError) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			msg, err := ReceiveMessage(ctx, p, url, 300, 20)
			if err != nil {
				// There was an error receiving the message
				e(err)
				continue
			}

			if msg == nil {
				// There were no messages in the queue, let's try again
				// No need to sleep, because internally the SDK does long-polling
				continue
			}

			err = f(msg)

			if err != nil {
				// If the callback function returns an error, leave the message in the queue
				e(err)

				// return the message back to the queue by reseting the visibility timeout
				if _, err := ChangeMsgVisibilityTimeout(ctx, p, url, msg.ReceiptHandle, 0); err != nil {
					e(err)
				}

				continue
			}

			err = DeleteMessage(msg.ReceiptHandle, p, url)
			if err != nil {
				// There was an error removing the message from the queue, so probably the message
				// is still in the queue and will receive it again (although we will never know),
				// so be prepared to process the message again without side effects.
				e(err)
				continue
			}
		}
	}
}