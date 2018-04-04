package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/jwt"
)

func Example_fromPrivateKey_signedToken() {
	// Get service JWK (in the real implementation, we'll read it from a file)
	retrieveJWK := func() ([]byte, error) {
		// generate a private key in JWK format, json encoded (like the one who generate the daemon)
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return []byte{}, err
		}

		pk, err := jwk.FromPrivateKey(privateKey)
		if err != nil {
			return []byte{}, err
		}

		return jwk.Marshal(pk)
	}

	jsonJWK, err := retrieveJWK()
	if err != nil {
		return
	}

	key, err := jwk.Unmarshal(jsonJWK)
	if err != nil {
		return
	}

	claims := &jwt.Claims{
		Sub: "life-service",
		Iss: "life.bs.laicosp.net",
	}

	signedtoken, err := jwt.NewSigned(jwt.SigningMethodRS256, key, claims)
	if err != nil {
		return
	}

	if len(strings.Split(signedtoken, ".")) != 3 {
		return
	}

	fmt.Println("token signed")
	// Output: token signed
}
