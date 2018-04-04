package jwk_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFilesystem_ReadWrite(t *testing.T) {
	assert := assert.New(t)
	path := filepath.Join(os.TempDir(), uuid.New())

	k1, err := jwk.NewTestKey(jwk.WithKid("k1"))
	assert.NoError(err)

	k2, err := jwk.NewTestKey(jwk.WithKid("k2"))
	assert.NoError(err)

	k3, err := jwk.NewTestKey(jwk.WithKid("k3"))
	assert.NoError(err)

	fs := jwk.NewFile(path)

	assert.NoError(fs.Write(k3, k2, k1))

	keys, err := fs.Read()
	assert.NoError(err)
	assert.Len(keys, 3)

	ks := jwk.NewKeySet(keys...)

	key, err := ks.Key("k1")
	assert.NotNil(key)
	assert.NoError(err)

	key, err = ks.Key("k2")
	assert.NotNil(key)
	assert.NoError(err)

	key, err = ks.Key("k3")
	assert.NotNil(key)
	assert.NoError(err)
}
