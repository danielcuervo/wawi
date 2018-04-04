package jwt

import (
	"context"
	"net/http"

	"github.com/socialpoint/bsk/pkg/httpc"
	"github.com/socialpoint/bsk/pkg/httpx"
	"github.com/socialpoint/bsk/pkg/jwk"
)

type contextKey int

const (
	// tokenKey is the key that holds the JWT token within a context
	tokenKey contextKey = iota
)

// ExtractToken returns the request token parsed and validated for convenient usage
func ExtractToken(r *http.Request) *Token {
	return ExtractTokenFromContext(r.Context())
}

// ExtractTokenFromContext returns the request token parsed and validated for convenient usage
func ExtractTokenFromContext(ctx context.Context) *Token {
	token := ctx.Value(tokenKey)

	if token != nil {
		return token.(*Token)
	}

	return nil
}

// InsertToken inserts given token into the context of the request, this is specially useful for testing
func InsertToken(r *http.Request, token *Token) *http.Request {
	ctx := context.WithValue(r.Context(), tokenKey, token)
	return r.WithContext(ctx)
}

// TokenValidator returns an adapter that validate JWT tokens coming in the header
func TokenValidator(keyFunc Keyfunc, constraints ...Constraint) httpx.Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := ParseFromRequestHeader(r)

			if err == ErrNoTokenInRequest {
				httpx.Respond(w, r, http.StatusUnauthorized, err.Error())
				return
			}

			if err != nil {
				httpx.Respond(w, r, http.StatusBadRequest, err.Error())

				return
			}

			if err := token.Validate(keyFunc, constraints...); err != nil {
				httpx.Respond(w, r, http.StatusUnauthorized, err.Error())

				return
			}

			h.ServeHTTP(w, InsertToken(r, token))
		})
	}
}

// AddHeader adds the auth header with the passed JWT token.
func AddHeader(token string) httpc.Decorator {
	return func(c httpc.Client) httpc.Client {
		return httpc.ClientFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Add("Authorization", "BEARER "+token)
			return c.Do(r)
		})
	}
}

// KeyProviderFunc returns a JWK when it is called
type KeyProviderFunc func() (*jwk.Key, error)

// SelfSignedHeader generates a JWT and adds it as auth header.
func SelfSignedHeader(f KeyProviderFunc, claims *Claims) httpc.Decorator {
	return func(c httpc.Client) httpc.Client {
		return httpc.ClientFunc(func(r *http.Request) (*http.Response, error) {
			key, err := f()
			if err != nil {
				return nil, err
			}

			token, err := NewSigned(SigningMethodRS256, key, claims)
			if err != nil {
				return nil, err
			}

			return AddHeader(token)(c).Do(r)
		})
	}
}
