package sk

import (
	"sort"
)

// Types of inputs
const (
	TypeBoolean   = "bool"
	TypeString    = "string"
	TypeStructure = "structure"
	TypeInteger   = "integer"
	TypeFloat     = "float"
	TypeFile      = "file"
)

// Catalog represents a catalog of services
type Catalog struct {
	Services   Services              `json:"services" hcl:"service"`
	Structures map[string]*Structure `json:"-" hcl:"structure"`
}

// Services represents a group of services
type Services map[string]*Service

// A Service is a collection of operations
type Service struct {
	Name        string     `json:"name"`
	Description string     `json:"description" hcl:"description"`
	Operations  Operations `json:"operations" hcl:"operation"`
}

// Operations represents a list of operations indexed by name
type Operations map[string]*Operation

// Operation represents a service operation
type Operation struct {
	Name        string             `json:"name"`
	Description string             `json:"description" hcl:"description"`
	Category    string             `json:"category" hcl:"category"`
	Method      string             `json:"method" hcl:"method"`
	Endpoint    string             `json:"endpoint" hcl:"endpoint"`
	Inputs      map[string]*Input  `json:"inputs,omitempty" hcl:"input,omitempty"`
	Outputs     map[string]*Output `json:"outputs,omitempty" hcl:"output,omitempty"`
	Errors      map[string]*Error  `json:"errors,omitempty" hcl:"errors,omitempty"`
}

// SortedInputs returns the operation inputs sorted by their name.
func (op *Operation) SortedInputs() SortedInputs {
	inputs := make(SortedInputs, len(op.Inputs))
	pos := 0
	for name, input := range op.Inputs {
		input.Name = name
		inputs[pos] = input
		pos++
	}

	sort.Sort(inputs)

	return inputs
}

// Input represents an input of an operation
type Input struct {
	Name        string        `json:"name"`
	Type        string        `json:"type" hcl:"type"`
	Required    bool          `json:"required,omitempty" hcl:"required,omitempty"`
	Description string        `json:"description,omitempty" hcl:"description,omitempty"`
	Enum        []interface{} `json:"enum,omitempty" hcl:"enum,omitempty"`
	Default     interface{}   `json:"default,omitempty" hcl:"default,omitempty"`
	Position    int           `json:"position,omitempty" hcl:"position,omitempty"`
}

// IsString return true when the input type is String
func (i *Input) IsString() bool {
	return i.Type == TypeString
}

// Output represents the output of an operation
type Output struct {
	Name        string `json:"name"`
	Type        string `json:"type" hcl:"type"`
	Description string `json:"description,omitempty" hcl:"description,omitempty"`
}

// Error represents a service error
type Error struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty" hcl:"description,omitempty"`
}

// Structure represents a complex structure used in inputs and outputs
type Structure struct {
}

// SortedInputs is a set of inputs sorted by position
type SortedInputs []*Input

func (si SortedInputs) Len() int {
	return len(si)
}

func (si SortedInputs) Less(i, j int) bool {
	return si[i].Position < si[j].Position
}

func (si SortedInputs) Swap(i, j int) {
	si[i], si[j] = si[j], si[i]
}
