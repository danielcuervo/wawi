package sk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/socialpoint/bsk/pkg/httpx"
	"github.com/socialpoint/bsk/pkg/sk"
)

func ExampleCatalogHandler_services_http() {
	catalogLoader := sk.NewStaticLoader(&sk.Catalog{})

	catalogHandler := httpx.AddHeaderDecorator("catalog", "example")(sk.CatalogHandler(catalogLoader))

	router := httpx.NewRouter()
	router.Route("/test", httpx.StatusOKHandler)
	router.Route(sk.CatalogPath, catalogHandler)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/_catalog", nil)
	if err != nil {
		panic(err)
	}

	router.ServeHTTP(w, r)
	fmt.Println(w.Body)

	// Output:
	// {"services":null}
}
