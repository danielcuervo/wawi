package jwt_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func TestCreateLtsTokenHTTPHandler(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()

	rsaPrivateKey := func() *rsa.PrivateKey {
		pk, err := rsa.GenerateKey(rand.Reader, jwk.DefaultRSABitSize)
		assert.NoError(err)

		return pk
	}()
	jwkPrivateKeyFunc := func(k *rsa.PrivateKey) *jwk.Key {
		pk, err := jwk.FromPrivateKey(k)
		assert.NoError(err)

		return pk
	}
	jwkPrivateKey := jwkPrivateKeyFunc(rsaPrivateKey)

	signedRequest := func() *http.Request {
		params := jwt.CreateLtsTokenRequestParams{
			ServiceName: "test-new-jwt-http-handler.bs.laicosp.net",
		}
		bodyRawContent, err := json.Marshal(params)
		assert.NoError(err)

		r, err := http.NewRequest("POST", "/api/jwt/create", bytes.NewReader(bodyRawContent))
		assert.NoError(err)

		claims := jwt.DefaultClaims()
		claims.Iss = "test.bs.laicosp.net"
		claims.Email = "test@socialpoint.es"
		claims.Exp = now.Add(time.Hour).Unix()

		requestToken, err := jwt.NewSigned(jwt.SigningMethodRS256, jwkPrivateKey, claims)
		assert.NoError(err)

		r.Header.Add("Authorization", "BEARER "+requestToken)
		requestTokenParsed, err := jwt.ParseFromRequestHeader(r)

		assert.NoError(err)
		r = jwt.InsertToken(r, requestTokenParsed)

		return r
	}()

	tokenHTTPHandler := jwt.CreateLtsTokenHTTPHandler(
		func(params *jwt.CreateLtsTokenRequestParams) (*jwk.Key, error) {
			return jwkPrivateKey, nil
		},
	)

	recorder := httptest.NewRecorder()
	tokenHTTPHandler.ServeHTTP(recorder, signedRequest)

	assert.Equal(http.StatusOK, recorder.Code)

	type Response struct {
		Token     string
		ExpiresAt int64
	}

	res := Response{}
	body := recorder.Body.String()

	err := json.Unmarshal([]byte(body), &res)
	assert.NoError(err)

	assert.True(len(res.Token) > 0)
	assert.True(res.ExpiresAt >= now.Add(time.Hour*24*90).Unix())
	assert.True(res.ExpiresAt <= now.Add(time.Hour*24*91).Unix())

	resTokenParsed, err := jwt.Parse(res.Token)
	assert.NoError(err)

	assert.Equal("test-new-jwt-http-handler.bs.laicosp.net", resTokenParsed.Claims.Iss)
	assert.Equal("test@socialpoint.es", resTokenParsed.Claims.Email)
	assert.True(resTokenParsed.Claims.Exp >= now.Add(time.Hour*24*90).Unix())
	assert.True(resTokenParsed.Claims.Exp <= now.Add(time.Hour*24*91).Unix())
}

func TestCreateLtsTokenHTTPHandlerWhenRequestTokenIsNil(t *testing.T) {
	assert := assert.New(t)

	tokenHTTPHandler := jwt.CreateLtsTokenHTTPHandler(
		func(params *jwt.CreateLtsTokenRequestParams) (*jwk.Key, error) {
			return nil, nil
		},
	)

	request, err := http.NewRequest("POST", "/api/jwt/create", nil)
	assert.NoError(err)

	recorder := httptest.NewRecorder()
	tokenHTTPHandler.ServeHTTP(recorder, request)

	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestCreateLtsTokenHTTPHandlerWhenKeyIsNil(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()

	rsaPrivateKey := func() *rsa.PrivateKey {
		pk, err := rsa.GenerateKey(rand.Reader, jwk.DefaultRSABitSize)
		assert.NoError(err)

		return pk
	}()
	jwkPrivateKeyFunc := func(k *rsa.PrivateKey) *jwk.Key {
		pk, err := jwk.FromPrivateKey(k)
		assert.NoError(err)

		return pk
	}
	jwkPrivateKey := jwkPrivateKeyFunc(rsaPrivateKey)

	signedRequest := func() *http.Request {
		params := jwt.CreateLtsTokenRequestParams{
			ServiceName: "test-new-jwt-http-handler.bs.laicosp.net",
		}
		bodyRawContent, err := json.Marshal(params)
		assert.NoError(err)

		r, err := http.NewRequest("POST", "/api/jwt/create", bytes.NewReader(bodyRawContent))
		assert.NoError(err)

		claims := jwt.DefaultClaims()
		claims.Iss = "test.bs.laicosp.net"
		claims.Email = "test@socialpoint.es"
		claims.Exp = now.Add(time.Hour).Unix()

		requestToken, err := jwt.NewSigned(jwt.SigningMethodRS256, jwkPrivateKey, claims)
		assert.NoError(err)

		r.Header.Add("Authorization", "BEARER "+requestToken)
		requestTokenParsed, err := jwt.ParseFromRequestHeader(r)

		assert.NoError(err)
		r = jwt.InsertToken(r, requestTokenParsed)

		return r
	}()

	tokenHTTPHandler := jwt.CreateLtsTokenHTTPHandler(
		func(params *jwt.CreateLtsTokenRequestParams) (*jwk.Key, error) {
			return nil, errors.New("Key not exists.")
		},
	)

	recorder := httptest.NewRecorder()
	tokenHTTPHandler.ServeHTTP(recorder, signedRequest)

	assert.Equal(http.StatusInternalServerError, recorder.Code)
}
