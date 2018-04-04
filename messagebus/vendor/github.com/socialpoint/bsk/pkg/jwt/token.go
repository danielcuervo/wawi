package jwt

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/socialpoint/bsk/pkg/uuid"
)

// TimeFunc provides the current time when parsing token to validate "exp" claim (expiration time).
// You can override it to use another time value.  This is useful for testing or if your
// server uses a different time zone than your tokens.
var TimeFunc = time.Now

// Keyfunc is to be used by parse methods use this callback function to supply
// the key for verification.  The function receives the parsed,
// but unverified Token.  This allows you to use propries in the
// Header of the token (such as `kid`) to identify which key to use.
type Keyfunc func(*Token) (interface{}, error)

// Token represents a JWT token. Different fields will be used depending on whether you're
// creating or parsing/verifying a token.
type Token struct {
	Header    *Header // The first segment of the token
	Claims    *Claims // The second segment of the token
	Signature string  // The third segment of the token
	Content   string  // Token content, the raw header and payload, used to verify signature
}

// Header represents the header part of a JWT token, as per the RFC.
// There are several header parameters in the specification, but we only include here those that are used.
type Header struct {
	// The "alg" (algorithm) header parameter identifies the cryptographic algorithm used to secure the JWT.
	// The processing of the "alg" header parameter, if present, requires that the value of the "alg" header
	// parameter MUST be one that is both supported and for which there exists a key for use with that algorithm
	// associated with the issuer of the JWT.
	// This header parameter is REQUIRED.
	Alg string

	// The typ (type) header parameter is used to declare that this data structure is a JWT. If a "typ" parameter
	// is present, it is RECOMMENDED that its value be "JWT".
	// This header parameter is OPTIONAL.
	Typ string

	// The key ID header parameter is a hint indicating which specific key owned by the signer should be
	// used to validate the signature. This allows signers to explicitly signal a change of key to recipients.
	// Omitting this parameter is equivalent to setting it to an empty string.
	// The interpretation of the contents of the "kid" parameter is unspecified.
	// This header parameter is OPTIONAL.
	Kid string

	// There are other header parameters in the specification, but we only use these for now.
}

// Validate returns an error if the header is not valid
func (header *Header) Validate() error {
	if header.Alg == "" {
		return ErrAlgUnspecified
	}

	if header.Alg != SigningMethodRS256.Alg() {
		return ErrAlgUnavailable
	}

	return nil
}

// Validate validates the token agains the given keyfunc and constrains
func (token *Token) Validate(keyFunc Keyfunc, constraints ...Constraint) error {

	if err := token.Header.Validate(); err != nil {
		return err
	}

	var err error
	// Lookup key
	var key interface{}
	if keyFunc == nil {
		// keyFunc was not provided.  short circuiting validation
		return NewValidationError("no Keyfunc was provided.", ValidationErrorUnverifiable)
	}

	if key, err = keyFunc(token); err != nil {
		// keyFunc returned an error
		return &ValidationError{Inner: err, Errors: ValidationErrorUnverifiable}
	}

	// Check expiration times
	vErr := &ValidationError{}
	now := TimeFunc().Unix()

	if now > token.Claims.Exp {
		vErr.Inner = fmt.Errorf("token is expired")
		vErr.Errors |= ValidationErrorExpired
	}

	if now < token.Claims.Nbf {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= ValidationErrorNotValidYet
	}

	// Validate constraints
	for _, constraint := range constraints {
		err := constraint(token)
		if err != nil {
			vErr.Inner = err
			vErr.Errors |= ValidationErrorUnsatisfiedConstraint
		}
	}

	// Perform validation
	if err = SigningMethodRS256.Verify(token.Content, token.Signature, key); err != nil {
		vErr.Inner = err
		vErr.Errors |= ValidationErrorSignatureInvalid
	}

	if vErr.valid() {
		return nil
	}

	return errors.New("invalid token")
}

// Claims contains the claims supported and enforced by this package
type Claims struct {
	Iss   string // Issuer who granted the JWT (e.g. "auth.service") as a StringOrURI.
	Sub   string // Subject of the JWT (often a user) StringOrURI which should either be a globally unique value or locally unique within a JWT communication context ("123121432423", "dghubble").
	Aud   string // Audience who should receive or consume the JWT, typically a JSON array of one or more StringOrURI values (e.g. ["mobile", "ios", "android"])
	Exp   int64  // Expiration time as a JSON numeric number of seconds since the epoch, after which the JWT will be rejected by library implementations
	Nbf   int64  // Similar to "exp", but the numeric time before which the JWT should not be accepted
	Iat   int64  // Issued at time which defines when the JWT was issued.
	Jti   string // A globally unique JWT token id string used to identify the JWT
	Hd    string // Hosted Domain claim
	Email string // The user's email address.
}

// New creates a new Token.  Takes a signing method
func New(method SigningMethod) *Token {
	return &Token{
		Header: &Header{
			Typ: "JWT",
			Alg: method.Alg(),
		},
		Claims: DefaultClaims(),
	}
}

// NewSigned create a new token and sign it with the given Key (Key is a JWK json encoded)
func NewSigned(m SigningMethod, k *jwk.Key, claims *Claims) (string, error) {
	pk, err := k.ToPrivateKey()
	if err != nil {
		return "", err
	}

	token := New(m)
	token.Claims = claims
	token.Header.Kid = k.Kid

	err = token.validateClaims()
	if err != nil {
		return "", err
	}

	return token.SignedString(pk)
}

// NewSignedTokenFromReader create a new token and sign it with the key provided by the reader
// The key is a JWK json encoded)
func NewSignedTokenFromReader(r io.Reader, m SigningMethod, claims *Claims) (string, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	key, err := jwk.Unmarshal(data)
	if err != nil {
		return "", err
	}
	return NewSigned(m, key, claims)
}

// SignedString returns the complete, signed token
func (token *Token) SignedString(key crypto.PrivateKey) (string, error) {
	var sig, sstr string
	var err error

	if sstr, err = token.SigningString(); err != nil {
		return "", err
	}

	if sig, err = SigningMethodRS256.Sign(sstr, key); err != nil {
		return "", err
	}

	return sstr + "." + sig, nil
}

// SigningString generates the signing string.  This is the
// most expensive part of the whole deal.  Unless you
// need this for something special, just go straight for
// the SignedString.
func (token *Token) SigningString() (string, error) {
	var err error
	var j []byte

	if j, err = json.Marshal(token.Header); err != nil {
		return "", err
	}

	sh := EncodeSegment(j)

	if j, err = json.Marshal(token.Claims); err != nil {
		return "", err
	}

	sc := EncodeSegment(j)

	return sh + "." + sc, nil
}

func (token *Token) validateClaims() error {
	if token.Claims.Iss == "" {
		return errors.New("Iss claim is empty")
	}
	return nil
}

// DefaultClaims returns the default claims for SP auth.
func DefaultClaims() *Claims {
	now := TimeFunc().Unix()

	return &Claims{
		Aud: "socialpoint.es",
		Hd:  "socialpoint.es",
		Nbf: now - 100,
		Exp: now + int64(time.Hour),
		Iat: now,
		Jti: uuid.New(),
	}
}

// ParseFromRequestHeader tries to find the token in an http.Request Authorization header.
func ParseFromRequestHeader(req *http.Request) (*Token, error) {
	// Look for an Authorization header
	if ah := req.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			return Parse(ah[7:])
		}
	}

	return nil, ErrNoTokenInRequest

}

// EncodeSegment encodes a JWT specific base64url encoding with padding stripped
func EncodeSegment(seg []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(seg), "=")
}

// DecodeSegment decodes a JWT specific base64url encoding with padding stripped
func DecodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}
