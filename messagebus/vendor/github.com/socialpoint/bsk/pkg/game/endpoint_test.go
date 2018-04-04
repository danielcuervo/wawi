package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointResolver(t *testing.T) {
	assert := assert.New(t)

	endpoint, err := DefaultEndpointResolver("dc", "fail")
	assert.Error(err)
	assert.Empty(endpoint)

	endpoint, err = DefaultEndpointResolver("dc", "android")
	assert.NoError(err)
	assert.Equal(endpoint, "http://dca.socialpointgames.com")

	endpoint, err = DefaultEndpointResolver("rc", "ios")
	assert.NoError(err)
	assert.Equal(endpoint, "https://rci.socialpointgames.com")
}
