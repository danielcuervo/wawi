package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/socialpoint/bsk/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	name        string
	claims      *jwt.Claims
	valid       bool
	errors      uint32
	constraints []jwt.Constraint
}

func makeSample(t *testing.T, claims *jwt.Claims) (string, jwt.Keyfunc) {
	key := generateTestKey(t)

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = claims

	s, err := token.SignedString(key)
	assert.NoError(t, err)

	kf := func(t *jwt.Token) (interface{}, error) {
		return key.Public(), nil
	}

	return s, kf
}

func TestParse(t *testing.T) {
	assert := assert.New(t)

	// Test cases generated using https://jwt.io/#debugger-io
	tests := []struct {
		str       string
		err       error
		alg       string
		content   string
		signature string
	}{
		{
			"a.b",
			jwt.ErrMalformedToken,
			"",
			"",
			"",
		},
		{
			"a.b.c",
			base64.CorruptInputError(0),
			"",
			"",
			"",
		},
		{
			"dyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIixiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE",
			&json.SyntaxError{},
			"RS256",
			"",
			"",
		},
		{
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIixiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE",
			&json.SyntaxError{},
			"RS256",
			"",
			"",
		},
		{
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE",
			nil,
			"RS256",
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9",
			"EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE",
		},
	}

	for _, tc := range tests {
		token, err := jwt.Parse(tc.str)
		assert.IsType(tc.err, err)

		if tc.err == nil {
			assert.Equal(tc.alg, token.Header.Alg)
			assert.Equal(tc.signature, token.Signature)
			assert.Equal(tc.content, token.Content)
		}
	}
}

func TestParseRequest(t *testing.T) {
	assert := assert.New(t)

	// Token generated using https://jwt.io/#debugger-io
	r1, _ := http.NewRequest("GET", "/", nil)
	r1.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE"))
	token, err := jwt.ParseFromRequestHeader(r1)
	assert.NotNil(token)
	assert.NoError(err)

	r2, _ := http.NewRequest("GET", "/", nil)
	r2.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "invalid"))
	token, err = jwt.ParseFromRequestHeader(r2)
	assert.Nil(token)
	assert.Error(err)

	r3, _ := http.NewRequest("GET", "/", nil)
	token, err = jwt.ParseFromRequestHeader(r3)
	assert.Nil(token)
	assert.Error(err)

}

// Helper method for benchmarking various methods
func benchmarkSigning(b *testing.B, method jwt.SigningMethod, key interface{}) {
	t := jwt.New(method)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := t.SignedString(key); err != nil {
				b.Fatal(err)
			}
		}
	})

}
