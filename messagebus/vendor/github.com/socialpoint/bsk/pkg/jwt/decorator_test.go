package jwt_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func TestInsertExtractToken(t *testing.T) {
	assert := assert.New(t)

	token1 := &jwt.Token{Claims: &jwt.Claims{Email: "demo@socialpoint.es"}}

	req := (&http.Request{}).WithContext(context.Background())

	req = jwt.InsertToken(req, token1)
	token2 := jwt.ExtractToken(req)
	token3 := jwt.ExtractTokenFromContext(req.Context())

	assert.EqualValues(token1, token2)
	assert.EqualValues(token1, token3)
}

func TestTokenValidator_ValidRequest(t *testing.T) {
	assert := assert.New(t)

	claims := &jwt.Claims{
		Nbf: time.Now().Unix() - 100,
		Exp: time.Now().Unix() + 1000,
		Iss: "test",
	}

	tokenStr, keyFunc := makeSample(t, claims)

	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Add("Authorization", "BEARER "+tokenStr)

	h := jwt.TokenValidator(keyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := jwt.ExtractToken(r)
		assert.NotNil(token)
		assert.Equal("test", token.Claims.Iss)
	}))

	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
}

func TestTokenValidator_Request_Without_Header(t *testing.T) {
	assert := assert.New(t)
	_, keyFunc := makeSample(t, nil)

	req, _ := http.NewRequest("GET", "", nil)

	h := jwt.TokenValidator(keyFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := jwt.ExtractToken(r)
		assert.Nil(token)
	}))

	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	assert.Equal(http.StatusUnauthorized, w.Code)
}

func TestSelfSignedHeader(t *testing.T) {
	assert := assert.New(t)

	f := func() (*jwk.Key, error) {
		pk, err := rsa.GenerateKey(rand.Reader, numBits)
		assert.NoError(err)

		key, err := jwk.FromPrivateKey(pk)
		assert.NoError(err)

		key.Alg = "RS256"
		key.Kid = "qwerty"

		return key, nil
	}

	claims := jwt.DefaultClaims()
	claims.Iss = "test"

	decorator := jwt.SelfSignedHeader(f, claims)
	client := decorator(&noopClient{})

	req, _ := http.NewRequest("GET", "", nil)
	resp, err := client.Do(req)

	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Regexp(`BEARER [\w_-]+\.[\w_-]+\.[\w_-]+`, req.Header.Get("Authorization"))
}

func makeToken(t *testing.T, claims *jwt.Claims) string {
	key := generateTestKey(t)

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = claims

	s, err := token.SignedString(key)
	assert.NoError(t, err)

	return s
}

type noopClient struct{}

func (c *noopClient) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}
