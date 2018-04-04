package jwk_test

import (
	"context"
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	assert := assert.New(t)

	k1, err := jwk.NewTestKey()
	assert.NoError(err)

	k2, err := jwk.NewTestKey()
	assert.NoError(err)

	src := jwk.NewKeySet(k1, k2)
	assert.NoError(err)

	dst := jwk.NewKeySet()

	err = jwk.Copy(dst, src)
	assert.NoError(err)

	assert.Contains(dst.Keys(), k1)
	assert.Contains(dst.Keys(), k2)
	assert.Contains(src.Keys(), k1)
	assert.Contains(src.Keys(), k2)
}

func TestLimitReader(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		total, limit int
	}{
		{10, 5},
		{10, 20},
		{10, 10},
		{10, 0},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			ks := jwk.NewKeySet()

			for i := 0; i < int(test.total); i++ {
				k, err := jwk.NewTestKey()
				assert.NoError(err)

				err = ks.Write(k)
				assert.NoError(err)
			}

			assert.Len(ks.Keys(), int(test.total))

			reader := jwk.LimitReader(ks, int64(test.limit))
			keys, err := reader.Read()
			assert.NoError(err)

			assert.True(len(keys) <= test.limit)

			if test.total < test.limit {
				assert.Equal(test.total, len(keys))
			}
		})
	}
}

func TestFilterExpiredReader(t *testing.T) {
	assert := assert.New(t)
	point := time.Now()

	expired, err := jwk.NewTestKey(jwk.WithNotAfter(point), jwk.WithKid("exp"))
	assert.NoError(err)

	active, err := jwk.NewTestKey(jwk.WithNotAfter(point.Add(2*time.Hour)), jwk.WithKid("active"))
	assert.NoError(err)

	ks := jwk.NewKeySet(expired, active)
	assert.Len(ks.Keys(), 2)

	filtered, err := jwk.FilterExpiredReader(ks, point.Add(time.Hour)).Read()
	assert.NoError(err)
	assert.Len(filtered, 1)

	assert.Contains(filtered, active)
	assert.NotContains(filtered, expired)
}

func TestCopyEvery(t *testing.T) {
	assert := assert.New(t)

	k1, err := jwk.NewTestKey(jwk.WithKid("k1"))
	assert.NoError(err)

	k2, err := jwk.NewTestKey(jwk.WithKid("k2"))
	assert.NoError(err)

	k3, err := jwk.NewTestKey(jwk.WithKid("k3"))
	assert.NoError(err)

	k4, err := jwk.NewTestKey(jwk.WithKid("k4"))
	assert.NoError(err)

	spy := &spyReadWriter{
		src: []*jwk.Key{k1, k2},
		dst: make(chan *jwk.Key),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jwk.CopyEvery(ctx, time.Microsecond, spy, spy)

	// After a while, write more keys
	time.AfterFunc(time.Microsecond, func() {
		err := spy.Write(k3, k4)
		assert.NoError(err)
	})

	// Receive the keys
	keys := []*jwk.Key{<-spy.dst, <-spy.dst, <-spy.dst, <-spy.dst}

	// Check that eventually all keys are copied
	kids := []string{keys[0].Kid, keys[1].Kid, keys[2].Kid, keys[3].Kid}
	assert.Contains(kids, k1.Kid)
	assert.Contains(kids, k2.Kid)
	assert.Contains(kids, k3.Kid)
	assert.Contains(kids, k4.Kid)
}

func TestTeeReader(t *testing.T) {
	assert := assert.New(t)

	k1, err := jwk.NewTestKey()
	assert.NoError(err)

	k2, err := jwk.NewTestKey()
	assert.NoError(err)

	src := jwk.NewKeySet(k1, k2)
	dst := jwk.NewKeySet()

	reader := jwk.TeeReader(src, dst)

	keys, err := reader.Read()
	assert.NoError(err)

	// Assert that keys were read and also replicated to the destination
	assert.Contains(keys, k1)
	assert.Contains(keys, k2)

	assert.Contains(dst.Keys(), k1)
	assert.Contains(dst.Keys(), k2)
}

func TestMultiReader(t *testing.T) {
	assert := assert.New(t)

	k1, err := jwk.NewTestKey()
	assert.NoError(err)

	k2, err := jwk.NewTestKey()
	assert.NoError(err)

	k3, err := jwk.NewTestKey()
	assert.NoError(err)

	k4, err := jwk.NewTestKey()
	assert.NoError(err)

	r1 := jwk.NewKeySet(k1, k2)
	r2 := jwk.NewKeySet(k3, k4)

	keys, err := jwk.MultiReader(r1, r2).Read()
	assert.NoError(err)

	assert.Len(keys, 4)
	assert.Contains(keys, k1)
	assert.Contains(keys, k2)
	assert.Contains(keys, k3)
	assert.Contains(keys, k4)
}

type spyReadWriter struct {
	src []*jwk.Key
	dst chan *jwk.Key
}

func (spy *spyReadWriter) Read() ([]*jwk.Key, error) {
	keys := make([]*jwk.Key, len(spy.src))
	copy(keys, spy.src)
	spy.src = nil
	return keys, nil
}

func (spy *spyReadWriter) Write(keys ...*jwk.Key) error {
	for _, k := range keys {
		spy.dst <- k
	}
	return nil
}
