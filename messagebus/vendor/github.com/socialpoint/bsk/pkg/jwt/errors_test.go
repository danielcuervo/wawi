package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	assert := assert.New(t)

	e := NewValidationError("", ValidationErrorExpired|ValidationErrorMalformed)

	assert.True(e.Is(ValidationErrorExpired))
	assert.True(e.Is(ValidationErrorMalformed))

	assert.False(e.Is(ValidationErrorNotValidYet))
	assert.False(e.Is(ValidationErrorSignatureInvalid))
}
