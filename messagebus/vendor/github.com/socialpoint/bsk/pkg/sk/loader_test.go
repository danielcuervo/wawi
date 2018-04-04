package sk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemLoader_Load(t *testing.T) {
	assert := assert.New(t)

	l, err := NewFilesystemLoader("fixtures/basic.hcl")
	assert.NoError(err)

	c := l.Load()

	assert.Equal("Basic service definition", c.Services["basic"].Description)
}
