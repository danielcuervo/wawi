package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/socialpoint/bsk/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

const numBits = 1024
const testMessage = "Life is beautiful but people are crazy."

func generateTestKey(t *testing.T) (key *rsa.PrivateKey) {
	assert := assert.New(t)

	key, err := rsa.GenerateKey(rand.Reader, numBits)
	assert.NoError(err)

	return
}

func TestRSASignAndVerify(t *testing.T) {
	assert := assert.New(t)
	key := generateTestKey(t)

	sig, err := jwt.SigningMethodRS256.Sign(testMessage, key)
	assert.NoError(err)
	assert.NotEmpty(sig)

	err = jwt.SigningMethodRS256.Verify(testMessage, sig, key.Public())
	assert.NoError(err)
}
