package timex_test

import (
	"context"
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/timex"
	"github.com/stretchr/testify/assert"
)

func TestRunInterval(t *testing.T) {
	assert := assert.New(t)

	ch := make(chan time.Time)

	f := func() {
		ch <- time.Now()
	}

	runner := timex.IntervalRunner(time.Millisecond, f)

	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)

	t1 := <-ch
	t2 := <-ch
	t3 := <-ch

	assert.WithinDuration(t2, t1, 3*time.Millisecond)
	assert.WithinDuration(t3, t2, 3*time.Millisecond)

	cancel()
}
