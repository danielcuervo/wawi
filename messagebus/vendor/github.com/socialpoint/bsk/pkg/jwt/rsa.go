package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

// SigningMethodRSA implements the RSA family of signing methods signing methods
type SigningMethodRSA struct {
	Name string
	Hash crypto.Hash
}

// Specific instances for RS256 and company
var (
	// RS256
	SigningMethodRS256 = &SigningMethodRSA{"RS256", crypto.SHA256}

	// RS384
	SigningMethodRS384 = &SigningMethodRSA{"RS384", crypto.SHA384}

	// RS512
	SigningMethodRS512 = &SigningMethodRSA{"RS512", crypto.SHA512}
)

// Alg implements the Alg method from SigningMethod
func (m *SigningMethodRSA) Alg() string {
	return m.Name
}

// Verify Implements the Verify method from SigningMethod
func (m *SigningMethodRSA) Verify(signingString, signature string, key crypto.PublicKey) error {
	var err error

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return ErrInvalidKey
	}

	// Decode the signature
	var sig []byte
	if sig, err = DecodeSegment(signature); err != nil {
		return err
	}

	// Create hasher
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}

	hasher := m.Hash.New()
	if _, err := hasher.Write([]byte(signingString)); err != nil {
		return err
	}

	// Verify the signature
	return rsa.VerifyPKCS1v15(rsaKey, m.Hash, hasher.Sum(nil), sig)
}

// Sign implements the Sign method from SigningMethod
func (m *SigningMethodRSA) Sign(signingString string, key crypto.PrivateKey) (string, error) {
	var err error

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", ErrInvalidKey
	}

	// Create the hasher
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := m.Hash.New()
	if _, err = hasher.Write([]byte(signingString)); err != nil {
		return "", err
	}

	// Sign the string and return the encoded bytes
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, m.Hash, hasher.Sum(nil))
	if err != nil {
		return "", err
	}

	return EncodeSegment(sigBytes), nil
}
