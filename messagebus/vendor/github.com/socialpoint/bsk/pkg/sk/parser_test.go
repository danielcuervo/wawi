package sk

import (
	"net/http"
	"os"
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
)

func Test_ParseFromHCL_BasicExample(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open("fixtures/basic.hcl")
	assert.NoError(err)
	dsl, err := ParseFromHCL(f)
	assert.NoError(err)

	basic, ok := dsl.Services["basic"]
	assert.True(ok)
	assert.Equal("basic", basic.Name)
	assert.Equal("Basic service definition", basic.Description)

	greet, ok := basic.Operations["greet"]
	assert.True(ok)
	assert.Equal(greet.Name, "greet")
	assert.Equal(greet.Description, "Greet by name")
	assert.Equal(greet.Method, http.MethodGet)

	name, ok := greet.Inputs["name"]
	assert.True(ok)
	assert.Equal(name.Name, "name")
	assert.Equal(name.Description, "the name")
	assert.Equal(name.Type, "string")
	assert.True(name.Required)
}

func Test_ParseFromHCL_MultipleExample(t *testing.T) {
	assert := assert.New(t)
	var ok bool

	f, err := os.Open("fixtures/multiple.hcl")
	assert.NoError(err)
	dsl, err := ParseFromHCL(f)
	assert.NoError(err)

	_, ok = dsl.Services["geoip"]
	assert.True(ok)

	_, ok = dsl.Services["crosspromotion"]
	assert.True(ok)

	get, ok := dsl.Services["crosspromotion"].Operations["get"]
	assert.True(ok)

	_, ok = get.Outputs["xpromo"]
	assert.True(ok)
}

func Test_ParseFromJSON_BasicExample(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open("fixtures/basic.json")
	assert.NoError(err)
	dsl, err := ParseFromJSON(f)
	assert.NoError(err)

	basic, ok := dsl.Services["basic"]
	assert.True(ok)
	assert.Equal("basic", basic.Name)
	assert.Equal("Basic service definition", basic.Description)

	greet, ok := basic.Operations["greet"]
	assert.True(ok)
	assert.Equal(greet.Name, "greet")
	assert.Equal(greet.Description, "Greet by name")
	assert.Equal(greet.Method, http.MethodGet)

	name, ok := greet.Inputs["name"]
	assert.True(ok)
	assert.Equal(name.Name, "name")
	assert.Equal(name.Description, "the name")
	assert.Equal(name.Type, "string")
	assert.True(name.Required)
}

func Test_ParseFromJSON_MultipleExample(t *testing.T) {
	assert := assert.New(t)
	var ok bool

	f, err := os.Open("fixtures/multiple.json")
	assert.NoError(err)
	dsl, err := ParseFromJSON(f)
	assert.NoError(err)

	_, ok = dsl.Services["geoip"]
	assert.True(ok)

	_, ok = dsl.Services["crosspromotion"]
	assert.True(ok)

	get, ok := dsl.Services["crosspromotion"].Operations["get"]
	assert.True(ok)

	_, ok = get.Outputs["xpromo"]
	assert.True(ok)
}

func Test_Invalid_Reader(t *testing.T) {
	assert := assert.New(t)

	c, err := ParseFromHCL(new(FailingReader))

	assert.Nil(c)
	assert.Error(err, "ooops")
}

type FailingReader struct{}

func (r *FailingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("ooops")
}
