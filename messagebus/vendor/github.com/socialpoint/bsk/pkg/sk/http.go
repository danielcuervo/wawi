package sk

import (
	"net/http"

	"github.com/socialpoint/bsk/pkg/httpx"
)

// CatalogPath is the default path recommended to services to expose their catalog at
const CatalogPath = "/_catalog"

// CatalogHandler returns a http.HandlerFunc that serves the content of a catalog loaded
// from the given loader
func CatalogHandler(loader Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalog := loader.Load()
		httpx.Respond(w, r, http.StatusOK, catalog)
	}
}
