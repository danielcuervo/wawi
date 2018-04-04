package jwk_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRSAPublicKeyMarshalingAndUnmarshaling(t *testing.T) {
	assert := assert.New(t)

	key, err := jwk.FromPublicKey(struct{}{})
	assert.Error(err)
	assert.Nil(key)

	pk, err := rsa.GenerateKey(rand.Reader, jwk.SimpleRSABitSize)
	assert.NoError(err)

	key, err = jwk.FromPublicKey(pk.Public())
	assert.NoError(err)

	key.Alg = "RS256"
	key.Kid = uuid.New()

	jsonKey, err := jwk.Marshal(key)
	assert.NoError(err)

	expKey, err := jwk.Unmarshal(jsonKey)
	assert.NoError(err)

	assert.EqualValues(expKey, key)
}

func TestPrivateKeyMarshalingAndUnmarshaling(t *testing.T) {
	assert := assert.New(t)

	key, err := jwk.FromPrivateKey(struct{}{})
	assert.Error(err)
	assert.Nil(key)

	pk, err := rsa.GenerateKey(rand.Reader, jwk.SimpleRSABitSize)
	assert.NoError(err)

	key, err = jwk.FromPrivateKey(pk)
	assert.NoError(err)

	key.Alg = "RS256"
	key.Kid = uuid.New()

	jsonKey, err := jwk.Marshal(key)
	assert.NoError(err)

	expKey, err := jwk.Unmarshal(jsonKey)
	assert.NoError(err)

	assert.Equal(expKey, key)

	privateKey, err := expKey.ToPrivateKey()
	assert.NoError(err)

	rsaPk := privateKey.(*rsa.PrivateKey)

	assert.Equal(pk.D, rsaPk.D)
	assert.Equal(pk.E, rsaPk.E)
	assert.Equal(pk.PublicKey, rsaPk.PublicKey)
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	// With kid & key
	k, err := jwk.NewTestKey(jwk.WithKid("1234"))
	assert.NoError(err)
	assert.Equal("1234", k.Kid)
	assert.NotEmpty(k.N)

	// With kid
	k, err = jwk.NewTestKey(jwk.WithKid("1234"))
	assert.NoError(err)
	assert.Equal("1234", k.Kid)
	assert.NotEmpty(k.N)

	// With key
	k, err = jwk.NewTestKey()
	assert.NoError(err)
	assert.NotEmpty(k.Kid)
	assert.NotEmpty(k.N)

	// With custom dates
	nb := time.Now().Add(5 * time.Minute)
	na := time.Now().Add(24 * time.Hour)
	k, err = jwk.NewTestKey(jwk.WithNotBefore(nb), jwk.WithNotAfter(na))
	assert.NoError(err)
	assert.Equal(nb, k.NotBefore)
	assert.Equal(na, k.NotAfter)

	// With nothing
	k, err = jwk.NewTestKey()
	assert.NoError(err)
	assert.NotEmpty(k.Kid)
	assert.NotEmpty(k.N)

	// With svc
	k, err = jwk.NewTestKey(jwk.WithSvc("life"))
	assert.NoError(err)
	assert.Equal("life", k.Svc)
}

func TestKey_Public(t *testing.T) {
	assert := assert.New(t)

	nb := time.Now()
	na := time.Now().Add(24 * time.Hour)
	k, err := jwk.NewTestKey(jwk.WithNotBefore(nb), jwk.WithNotAfter(na))
	assert.NoError(err)

	public := k.Public().(*jwk.Key)
	assert.NotEmpty(public.Kid)
	assert.Equal("RSA", public.Kty)
	assert.Equal("RS256", public.Alg)
	assert.NotEmpty(public.N)
	assert.NotEmpty(public.E)
	assert.Empty(public.D)
	assert.Equal(nb, public.NotBefore)
	assert.Equal(na, public.NotAfter)
}

func TestKey_IsActive(t *testing.T) {
	assert := assert.New(t)

	key, err := jwk.NewTestKey(
		jwk.WithNotBefore(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
		jwk.WithNotAfter(time.Date(2016, 12, 1, 0, 0, 0, 0, time.UTC)),
	)
	assert.NoError(err)

	for _, testcase := range []struct {
		time     time.Time
		expected bool
	}{
		{time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC), false},  // past
		{time.Date(2016, 6, 1, 0, 0, 0, 0, time.UTC), true},   // present
		{time.Date(2017, 6, 1, 0, 0, 0, 0, time.UTC), false},  // future
		{time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC), false},  // edge
		{time.Date(2016, 12, 1, 0, 0, 0, 0, time.UTC), false}, // edge
	} {
		assert.Equal(testcase.expected, key.IsActive(testcase.time))
	}
}

func TestKey_Zero_Value_Is_Always_Active(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()

	key, err := jwk.NewTestKey()
	assert.NoError(err)

	for _, t := range []time.Time{
		now, // present

		now.Add(-time.Hour),        // near past
		now.Add(time.Hour * 10000), //distant  past

		now.Add(time.Hour),         // future
		now.Add(time.Hour * 10000), // distant future

	} {
		assert.True(key.IsActive(t))
	}
}
