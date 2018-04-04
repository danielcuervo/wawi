// +build integration

package redisc_test

import (
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/redisc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	errorMsg   = "randomError"
	eventsList = "sp-integration-tests"
)

var payload = []byte("payload")

func TestNewRedisClientWithWrongPort(t *testing.T) {
	client, err := redisc.NewRedisClient("localhost:1234")
	require.NoError(t, err)

	assert.Error(t, client.Push(eventsList, payload))
}

func TestPushLocalhost(t *testing.T) {
	client, err := redisc.NewRedisClient("127.0.0.1:6379")
	require.NoError(t, err)
	defer client.Close()

	//WHEN
	require.NoError(t, client.Push(eventsList, payload))
	time.Sleep(100 * time.Millisecond)
	reply, err := client.Pop(eventsList)
	require.NoError(t, err)

	//THEN
	assert.Equal(t, payload, reply)
}
