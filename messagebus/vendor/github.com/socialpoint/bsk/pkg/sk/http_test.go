package sk

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPEndpoint(t *testing.T) {
	assert := assert.New(t)

	loader, err := NewFilesystemLoader("fixtures/basic.hcl")
	assert.NoError(err)

	catalog := loader.Load()
	encodedCatalog, err := json.Marshal(catalog)
	assert.NoError(err)
	encodedCatalog = append(encodedCatalog, '\n')

	req, err := http.NewRequest("GET", "/_catalog/", nil)
	recorder := httptest.NewRecorder()

	CatalogHandler(loader)(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)
	response, err := ioutil.ReadAll(recorder.Body)
	assert.NoError(err)

	// assert it serves the content of the catalog
	assert.Equal(encodedCatalog, response)
}
