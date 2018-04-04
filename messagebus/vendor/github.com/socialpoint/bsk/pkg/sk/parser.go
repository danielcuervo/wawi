package sk

import (
	"io"
	"io/ioutil"

	"encoding/json"

	"github.com/hashicorp/hcl"
)

// A Parser is a function that from a reader returns a Catalog
type Parser func(r io.Reader) (*Catalog, error)

// ParseFromHCL parses an HCL and returns a Catalog
func ParseFromHCL(r io.Reader) (*Catalog, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var c Catalog
	err = hcl.Unmarshal(content, &c)

	assignNames(&c)

	return &c, err
}

// ParseFromJSON parses a JSON and returns a Catalog
func ParseFromJSON(r io.Reader) (*Catalog, error) {
	var c Catalog
	err := json.NewDecoder(r).Decode(&c)

	assignNames(&c)

	return &c, err
}

func assignNames(c *Catalog) {
	for sName, service := range c.Services {
		service.Name = sName

		for opName, op := range service.Operations {
			op.Name = opName

			for iName, input := range op.Inputs {
				input.Name = iName
			}

			for oName, output := range op.Outputs {
				output.Name = oName
			}

			for eName, err := range op.Errors {
				err.Name = eName
			}
		}
	}
}
