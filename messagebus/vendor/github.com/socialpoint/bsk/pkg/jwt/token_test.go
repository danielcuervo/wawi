package jwt_test

import (
	"strings"
	"testing"

	"github.com/socialpoint/bsk/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSignedTokenFromFile(t *testing.T) {
	assert := assert.New(t)

	fakeJWKfile := `{"kty":"RSA","kid":"5786503a-83da-b89a-cb42-332fb5be6db1","alg":"RS256","d":"GQ9AfsYhJrUeEW0_30eFRAgkE_rOJUjpMN3lwjl-Hxw0OR-Q6XSKLR_LqlPFCc0rAohdZReHjP8H18pQHtRACGQqC5krFzZXJ--oWX2jb-zG9Mg8LZ5UYbJQebidGsptNDsuEHXcApgNbRrbl0B9JXA1Yt3REzaFUi8LmliaKFEip52ZN3VL9YzLUGGTPAI3TYYUwTzNRwQEBM3G2xBPzI1ZyCWkJyDN7PpFlr31QroL483tOGmOC4uzlE-Mg84PRe-1BPXqfwOGDWZA-QLM-a2AcozeP5Awoy4LVA3SlVG-f1C6AIuURYDn-s-KF3c9NQ56TCox1fzlr1jx3KEoeQ","n":"tlwHcr44pbpvukgCoVNqeBv5S4XxyjEwi19p6rdpP5GTkOg8vvJm8pUV8q5JfwLPCvRX7tt2jJvoxXKqeMMUipu6LiQBmkcEcTakVP3sQqKs_CsBs4zDVzqs0_iwj-7xzlK6AZbz51QVS5AjFIaFUEFYYIKkVCDC57wrs3wBnjPsSNEJZZ8LTB_UQYtHBFG4z_ZEvJTWZnR8sGocobcRt5YQssLPglu48T8uNjo-g_kaj32iQh-OQPYqQDIKi2AmeR6IWjfwABt9oYUVZBJGMyH32RU4lEqRFmiSVmqDaLZ076S65dC_SQ-1UomRpGk2atDVZZXDa4JmU2sULIpMsQ","e":"AQAB"}`

	reader := strings.NewReader(fakeJWKfile)

	claims := jwt.DefaultClaims()
	claims.Iss = "test.issuer"

	signedtoken, err := jwt.NewSignedTokenFromReader(reader, jwt.SigningMethodRS256, claims)
	assert.NoError(err)
	assert.Len(strings.Split(signedtoken, "."), 3)
}

func TestHeader_Validate(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		header *jwt.Header
		err    error
	}{
		{
			&jwt.Header{},
			jwt.ErrAlgUnspecified,
		},
		{
			&jwt.Header{Alg: "luissss"},
			jwt.ErrAlgUnavailable,
		},
		{
			&jwt.Header{Alg: "RS256"},
			nil,
		},
	}

	for _, tc := range tests {
		err := tc.header.Validate()
		assert.IsType(tc.err, err)
	}
}
