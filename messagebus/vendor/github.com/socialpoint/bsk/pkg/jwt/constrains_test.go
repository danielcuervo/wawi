package jwt_test

import (
	"testing"

	"github.com/socialpoint/bsk/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func newTestToken(claims *jwt.Claims) *jwt.Token {
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = claims

	return token
}

func TestHD(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		hd    string
		token *jwt.Token
		err   error
	}{
		{
			"socialpoint.es",
			newTestToken(&jwt.Claims{Hd: "socialpoint.es"}),
			nil,
		},
		{
			"example.com",
			newTestToken(&jwt.Claims{Hd: "socialpoint.es"}),
			jwt.ConstraintError(""),
		},
	}

	for _, tc := range tests {
		err := jwt.HD(tc.hd)(tc.token)
		assert.IsType(tc.err, err)
	}
}

func TestIss(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		iss   []string
		token *jwt.Token
		err   error
	}{
		{
			[]string{"socialpoint.es"},
			newTestToken(&jwt.Claims{Iss: "socialpoint.es"}),
			nil,
		},
		{
			[]string{"example.com", "socialpoint.es"},
			newTestToken(&jwt.Claims{Iss: "socialpoint.es"}),
			nil,
		},
		{
			[]string{"example.com", "socialpoint.es"},
			newTestToken(&jwt.Claims{Iss: "invalid.es"}),
			jwt.ConstraintError(""),
		},
	}

	for _, tc := range tests {
		err := jwt.Iss(tc.iss...)(tc.token)
		assert.IsType(tc.err, err)
	}
}

func TestAud(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		aud   string
		token *jwt.Token
		err   error
	}{
		{
			"socialpoint.es",
			newTestToken(&jwt.Claims{Aud: "socialpoint.es"}),
			nil,
		},
		{
			"example.com",
			newTestToken(&jwt.Claims{Aud: "socialpoint.es"}),
			jwt.ConstraintError(""),
		},
	}

	for _, tc := range tests {
		err := jwt.Aud(tc.aud)(tc.token)
		assert.IsType(tc.err, err)
	}
}
