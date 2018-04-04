package jwk

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/socialpoint/bsk/pkg/uuid"
)

const (
	// DefaultRSABitSize is the bit size to be used when generating RSA keys.
	// This default size or bigger is the recommended size for production purposes.
	DefaultRSABitSize = 2048

	// SimpleRSABitSize is a really low bit size for RSA keys, but is the recommended size
	// for testing purposes since it doesn't have an impact on performance.
	SimpleRSABitSize = 128
)

// A Source is anything that can return a Key.
type Source interface {
	Key(kid string) (*Key, error)
}

// Key represents the JSON data structure that represents a JSON Web Key (JWK).
// RFC: https://tools.ietf.org/html/rfc7517
type Key struct {
	// The fields below indicate common information, relative to nature
	// of the cryptographic key.
	Kty string `json:"kty,omitempty"`
	Use string `json:"use,omitempty"`
	Kid string `json:"kid,omitempty"`
	Alg string `json:"alg,omitempty"`

	// The fields below belong to keys based on RSA algorithm.
	D string `json:"d,omitempty"`
	N string `json:"n,omitempty"`
	E string `json:"e,omitempty"`

	// The fields below are not part of the RFC.
	// They have been added to support temporary behaviour such as the valid
	// period of a key.
	// This behaviour can be obtained creating a X509 Certificate, but this would add more complexity:
	// https://golang.org/pkg/crypto/x509/#CreateCertificate
	//
	// For simplicity we are including them as part of the Key.
	NotBefore time.Time `json:"nb,omitempty"`
	NotAfter  time.Time `json:"na,omitempty"`

	// The field below is not part of the RFC. It's used for monitoring and introspection.
	Svc string `json:"svc,omitempty"`
}

// Key returns the key itself if the kid match, this way Key implements the Source interface
func (key *Key) Key(kid string) (*Key, error) {
	if key.Kid == kid {
		return key, nil
	}

	return nil, nil
}

// IsActive returns whether the key is active or not in the given time
func (key *Key) IsActive(t time.Time) bool {
	return key.NotBefore.Before(t) && (key.NotAfter.After(t) || key.NotAfter.IsZero())
}

type options struct {
	kid       string
	key       *rsa.PrivateKey
	notBefore time.Time
	notAfter  time.Time
	svc       string
}

// Option configures a factory parameter
type Option func(*options)

// WithKid is an Option that sets the Kid you want to apply to the factory
func WithKid(kid string) Option {
	return func(opts *options) {
		opts.kid = kid
	}
}

// WithKey is an Option that sets the Key you want to apply to the factory
func WithKey(key *rsa.PrivateKey) Option {
	return func(opts *options) {
		opts.key = key
	}
}

// WithNotBefore is an Option that sets the notBefore you want to apply to the factory
func WithNotBefore(nb time.Time) Option {
	return func(opts *options) {
		opts.notBefore = nb
	}
}

// WithNotAfter is an Option that sets the notAfter you want to apply to the factory
func WithNotAfter(na time.Time) Option {
	return func(opts *options) {
		opts.notAfter = na
	}
}

// WithSvc is an Option that sets the svc field of the key
func WithSvc(svc string) Option {
	return func(opts *options) {
		opts.svc = svc
	}
}

// New creates a new JWK
func New(opts ...Option) (*Key, error) {
	options := &options{}
	for _, o := range opts {
		o(options)
	}

	if options.kid == "" {
		options.kid = uuid.New()
	}

	if options.key == nil {
		key, err := rsa.GenerateKey(rand.Reader, DefaultRSABitSize)
		if err != nil {
			return nil, err
		}

		options.key = key
	}

	jwk, err := FromPrivateKey(options.key)
	if err != nil {
		return nil, err
	}

	jwk.Kid = options.kid
	jwk.NotBefore = options.notBefore
	jwk.NotAfter = options.notAfter
	jwk.Svc = options.svc

	return jwk, nil
}

// NewTestKey creates a new JWK suitable for testing, using low bit size for RSA keys.
// Using a low bit size is not recommended for production usage,
// but for tests it's more performant.
func NewTestKey(opts ...Option) (*Key, error) {
	pk, err := rsa.GenerateKey(rand.Reader, SimpleRSABitSize)
	if err != nil {
		return nil, err
	}

	opts = append(opts, WithKey(pk))

	return New(opts...)
}

// Unmarshal  is a wrapper to unmarshal a JSON octet stream to a structured JWK
func Unmarshal(jwt []byte) (key *Key, err error) {
	key = &Key{}
	err = json.Unmarshal(jwt, key)
	return
}

// Marshal is a wrapper to marshal a JSON octet stream from a structured JWK
func Marshal(key *Key) ([]byte, error) {
	return json.Marshal(key)
}

// FromPublicKey create a JWK from a public key
func FromPublicKey(k crypto.PublicKey) (*Key, error) {
	key, ok := k.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Unknown key type %T", key)
	}

	jwt := &Key{
		Kty:       "RSA",
		Alg:       "RS256",
		N:         safeEncode(key.N.Bytes()),
		E:         safeEncode(big.NewInt(int64(key.E)).Bytes()),
		NotBefore: time.Time{}.In(time.UTC),
		NotAfter:  time.Time{}.In(time.UTC),
	}

	return jwt, nil
}

// ToPublicKey decodes a Key structure as a public key
func (key *Key) ToPublicKey() (crypto.PublicKey, error) {
	if key.Kty != "RSA" {
		return nil, fmt.Errorf("Unknown JWK key type %s", key.Kty)
	}

	if key.N == "" || key.E == "" {
		return nil, errors.New("Malformed JWK RSA key")
	}

	// decode exponent
	data, err := safeDecode(key.E)
	if err != nil {
		return nil, errors.New("Malformed JWK RSA key")
	}
	if len(data) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(data):], data)
		data = ndata
	}

	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(data[:])),
	}

	data, err = safeDecode(key.N)
	if err != nil {
		return nil, errors.New("Malformed JWK RSA key")
	}
	pubKey.N.SetBytes(data)

	return pubKey, nil
}

// Public returns the public key corresponding to key.
func (key *Key) Public() crypto.PublicKey {
	return &Key{
		Kid:       key.Kid,
		Kty:       "RSA",
		Alg:       "RS256",
		N:         key.N,
		E:         key.E,
		NotBefore: key.NotBefore,
		NotAfter:  key.NotAfter,
		Svc:       key.Svc,
	}
}

// FromPrivateKey creates a JWK from a private key
func FromPrivateKey(k crypto.PrivateKey) (*Key, error) {
	key, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("Unknown key type %T", key)
	}

	j, err := FromPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	j.D = safeEncode(key.D.Bytes())

	return j, err
}

// ToPrivateKey decodes a Key structure as a private key
func (key *Key) ToPrivateKey() (crypto.PrivateKey, error) {
	if key.Kty != "RSA" {
		return nil, fmt.Errorf("Unknown JWK key type %s", key.Kty)
	}

	if key.D == "" {
		return nil, errors.New("Malformed JWK RSA key")
	}

	pu, err := key.ToPublicKey()
	if err != nil {
		return nil, err
	}

	data, err := safeDecode(key.D)
	if err != nil {
		return nil, errors.New("Malformed JWK RSA key")
	}

	privKey := &rsa.PrivateKey{
		PublicKey: *pu.(*rsa.PublicKey),
		D:         &big.Int{},
	}

	privKey.D.SetBytes(data)

	return privKey, nil
}

func safeDecode(str string) ([]byte, error) {
	lenMod4 := len(str) % 4
	if lenMod4 > 0 {
		str = str + strings.Repeat("=", 4-lenMod4)
	}

	return base64.URLEncoding.DecodeString(str)
}

func safeEncode(p []byte) string {
	data := base64.URLEncoding.EncodeToString(p)
	return strings.TrimRight(data, "=")
}
