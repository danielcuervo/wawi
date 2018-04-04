package jwt

import "crypto"

// SigningMethod represents methods of signing or verifying tokens.
type SigningMethod interface {
	// Verify returns nil if signature is valid
	Verify(signing, signature string, key crypto.PublicKey) error

	// Sign returns encoded signature or error
	Sign(signing string, key crypto.PrivateKey) (string, error)

	// Alg returns the alg identifier for the method (example: 'HS256')
	Alg() string
}
