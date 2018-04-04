package jwt

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/socialpoint/bsk/pkg/httpx"
	"github.com/socialpoint/bsk/pkg/jwk"
)

const (
	tokenExpiration = time.Hour * 24 * 90
)

// CreateLtsTokenRequestParams represents request parameters
type CreateLtsTokenRequestParams struct {
	ServiceName string
}

// CreateLtsTokenKeyProviderFunc returns a function that provides a key
type CreateLtsTokenKeyProviderFunc func(*CreateLtsTokenRequestParams) (*jwk.Key, error)

// CreateLtsTokenHTTPHandler returns a http handler that creates JWT valid for 90 days
func CreateLtsTokenHTTPHandler(kpf CreateLtsTokenKeyProviderFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestToken := ExtractToken(r)
		if requestToken == nil {
			httpx.Respond(w, r, http.StatusBadRequest, errors.New("Request token not exists"))
			return
		}

		params := CreateLtsTokenRequestParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			httpx.Respond(w, r, http.StatusBadRequest, err.Error())
			return
		}

		key, err := kpf(&params)
		if err != nil {
			httpx.Respond(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		claims := func() *Claims {
			claims := DefaultClaims()
			claims.Iss = params.ServiceName
			claims.Email = requestToken.Claims.Email
			claims.Exp = time.Now().Add(tokenExpiration).Unix()

			return claims
		}()

		generatedToken, err := NewSigned(SigningMethodRS256, key, claims)
		if err != nil {
			httpx.Respond(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		httpx.Respond(w, r, http.StatusOK, struct {
			Token     string
			ExpiresAt int64
		}{
			Token:     generatedToken,
			ExpiresAt: claims.Exp,
		})
	})
}
