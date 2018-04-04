package jwk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/stretchr/testify/assert"
)

func TestNewKeySet(t *testing.T) {
	assert := assert.New(t)

	ks := jwk.NewKeySet()
	assert.IsType((*jwk.KeySet)(nil), ks)
}

func TestKeySet_Key(t *testing.T) {
	assert := assert.New(t)

	k1 := createKey(t, "1234")

	ks := jwk.NewKeySet(k1)

	// Valid kid
	key, err := ks.Key(k1.Kid)
	assert.NoError(err)
	assert.NotNil(key)

	// Invalid kid
	key, err = ks.Key("not_valid")
	assert.Error(err)
	assert.Nil(key)
}

func TestKeySet_Keys(t *testing.T) {
	assert := assert.New(t)

	k1 := createKey(t, "1234")
	k2 := createKey(t, "5678")

	ks := jwk.NewKeySet(k1, k2)
	keys := ks.Keys()
	assert.Len(keys, 2)

	assert.Contains(keys, k1)
	assert.Contains(keys, k2)
}

func TestKeySet_Write(t *testing.T) {
	assert := assert.New(t)

	k1 := createKey(t, "1234")
	k2 := createKey(t, "5678")

	ks := jwk.NewKeySet()
	err := ks.Write(k1, k2)
	assert.NoError(err)

	key, err := ks.Key(k1.Kid)
	assert.NoError(err)
	assert.NotNil(key)

	key, err = ks.Key(k2.Kid)
	assert.NoError(err)
	assert.NotNil(key)
}

func TestKeySet_Reset(t *testing.T) {
	assert := assert.New(t)

	k1 := createKey(t, "1234")
	k2 := createKey(t, "5678")

	ks := jwk.NewKeySet()
	err := ks.Write(k1, k2)
	assert.NoError(err)

	keys := ks.Keys()
	assert.Len(keys, 2)

	ks.Reset()

	keys = ks.Keys()
	assert.Len(keys, 0)
}

func TestKeySet_Marshal_And_UnMarshal_JSON(t *testing.T) {
	assert := assert.New(t)

	k1 := createKey(t, "1234")
	k2 := createKey(t, "5678")

	ks := jwk.NewKeySet(k1, k2)

	j, err := json.Marshal(ks)
	assert.NoError(err)
	assert.NotNil(j)

	var ksu jwk.KeySet
	err = json.Unmarshal(j, &ksu)
	assert.NoError(err)

	ku1, err := ksu.Key(k1.Kid)
	assert.NoError(err)
	assert.EqualValues(k1, ku1)

	ku2, err := ksu.Key(k2.Kid)
	assert.NoError(err)
	assert.EqualValues(k2, ku2)
}

func TestKeySet_MarshalJSON(t *testing.T) {
	assert := assert.New(t)

	k1 := createKey(t, "1234")

	ks := jwk.NewKeySet(k1)

	res, err := json.Marshal(ks)
	assert.NoError(err)

	var unserialized map[string][]*jwk.Key
	err = json.Unmarshal(res, &unserialized)
	assert.NoError(err)

	keys, ok := unserialized["keys"]
	assert.True(ok)
	assert.Len(keys, 1)
	assert.EqualValues(k1, keys[0])
}

func TestKeySet_UnMarshalJSON(t *testing.T) {
	assert := assert.New(t)

	for _, test := range []struct {
		payload []byte
		len     int
		err     error
	}{
		{[]byte(`{"keys":[{"kid":"1234"},{"kid":"5678"}]}`), 2, nil},
		{[]byte(`{"Keys":[{"kid":"1234"},{"kid":"5678"}]}`), 2, nil},
		{[]byte(`{"KEYS":[{"kid":"1234"},{"kid":"5678"}]}`), 2, nil},
		{[]byte(`[{"kid":"1234"},{"kid":"5678"}]`), 0, &json.UnmarshalTypeError{}},
		{[]byte(`]invalid_json[`), 0, &json.SyntaxError{}},
	} {
		var ks jwk.KeySet
		err := json.Unmarshal(test.payload, &ks)

		assert.IsType(test.err, err)
		assert.Len(ks.Keys(), test.len)
	}
}

func createKey(t *testing.T, kid string) *jwk.Key {
	k, err := jwk.NewTestKey(
		jwk.WithKid(kid),
		jwk.WithNotBefore(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
		jwk.WithNotAfter(time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)),
	)

	if err != nil {
		t.Fail()
	}

	return k
}

func TestSortKeysByNotBefore(t *testing.T) {
	assert := assert.New(t)

	k1, err := jwk.NewTestKey(jwk.WithKid("k1"), jwk.WithNotBefore(time.Now().Add(2*time.Hour)))
	assert.NoError(err)

	k2, err := jwk.NewTestKey(jwk.WithKid("k2"), jwk.WithNotBefore(time.Now().Add(4*time.Hour)))
	assert.NoError(err)

	k3, err := jwk.NewTestKey(jwk.WithKid("k3"), jwk.WithNotBefore(time.Now().Add(6*time.Hour)))
	assert.NoError(err)

	keys := []*jwk.Key{k2, k3, k1}
	jwk.SortKeysByNotBefore(keys)

	assert.Equal(k1, keys[0])
	assert.Equal(k2, keys[1])
	assert.Equal(k3, keys[2])
}

func TestKeySet_ServeHTTP(t *testing.T) {
	assert := assert.New(t)

	key, err := jwk.NewTestKey()
	assert.NoError(err)

	w := httptest.NewRecorder()
	r := &http.Request{}

	ks := jwk.NewKeySet(key)

	ks.ServeHTTP(w, r)

	assert.Equal(http.StatusOK, w.Code)
	assert.NotEmpty(w.Body)
}

func TestKeySet_Expire(t *testing.T) {
	assert := assert.New(t)
	point := time.Now()

	ek, err := jwk.NewTestKey(jwk.WithNotAfter(point), jwk.WithKid("expired"))
	assert.NoError(err)

	ak, err := jwk.NewTestKey(jwk.WithNotAfter(point.Add(2*time.Hour)), jwk.WithKid("active"))
	assert.NoError(err)

	ks := jwk.NewKeySet(ek, ak)
	assert.Len(ks.Keys(), 2)

	ks.Expire(point.Add(time.Minute))
	assert.Len(ks.Keys(), 1)

	assert.NotNil(ks.Key("active"))
	assert.Nil(ks.Key("expired"))
}

func TestKeySet_ActiveKey(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()

	assert.Nil(jwk.NewKeySet().ActiveKey(now), "An empty key set does not contains active keys")

	expired, err := jwk.NewTestKey(jwk.WithNotAfter(now.Add(-time.Hour)))
	assert.NoError(err)
	assert.Nil(jwk.NewKeySet(expired).ActiveKey(now), "A key set with an expired key does not contains active keys")

	active, err := jwk.NewTestKey(jwk.WithNotBefore(now.Add(-time.Hour)), jwk.WithNotAfter(now.Add(time.Hour)))
	assert.NoError(err)
	assert.EqualValues(active.Kid, jwk.NewKeySet(active).ActiveKey(now).Kid, "a key set with an active key should return the key")
}
