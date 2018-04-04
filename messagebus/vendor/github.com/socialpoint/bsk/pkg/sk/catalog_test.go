package sk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Sorted_Inputs(t *testing.T) {
	assert := assert.New(t)

	operation := Operation{
		Inputs: map[string]*Input{
			"bar": &Input{
				Position: 2,
			},
			"baz": &Input{
				Position: 3,
			},
			"foo": &Input{
				Position: 1,
			},
		},
	}

	inputs := operation.SortedInputs()
	assert.Equal("foo", inputs[0].Name)
	assert.Equal("bar", inputs[1].Name)
	assert.Equal("baz", inputs[2].Name)
}

func Test_Input_IsString(t *testing.T) {
	assert := assert.New(t)

	for _, d := range []struct {
		inputType string
		expected  bool
	}{
		{TypeString, true},
		{TypeBoolean, false},
		{TypeInteger, false},
		{TypeFloat, false},
		{TypeFile, false},
		{TypeStructure, false},
	} {
		input := Input{Type: d.inputType}
		assert.Equal(d.expected, input.IsString())
	}
}
