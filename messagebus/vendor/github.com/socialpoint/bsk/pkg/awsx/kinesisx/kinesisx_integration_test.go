// +build integration

package kinesisx_test

import (
	"testing"

	"github.com/socialpoint/bsk/pkg/awsx/kinesisx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	iterations = 3
)

func TestPopEmptyStream(t *testing.T) {
	kin := kinesisx.NewKinesisForTest(t)

	//WHEN
	popped, err := kin.Pop()

	//THEN
	assert.NoError(t, err)
	assert.Nil(t, popped)
}

func Test_PushNPopN(t *testing.T) {
	//GIVEN
	kin := kinesisx.NewKinesisForTest(t)

	//WHEN
	for i := 0; i < iterations; i++ {
		push(t, kin, kinesisx.GetTestMessage(i))
	}

	//THEN
	for i := 0; i < iterations; i++ {
		popped, err := kin.Pop()
		assert.NoError(t, err)
		assert.Equal(t, kinesisx.GetTestMessage(i), string(popped))
	}
}

func Test_PushPopN(t *testing.T) {
	//GIVEN
	kin := kinesisx.NewKinesisForTest(t)

	for i := 0; i < iterations; i++ {
		//WHEN
		msg := kinesisx.GetTestMessage(i)
		push(t, kin, msg)

		//THEN
		popped, err := kin.Pop()
		assert.NoError(t, err)
		assert.Equal(t, msg, string(popped))
	}
}

func push(t *testing.T, kin kinesisx.StreamWriter, msg string) {
	require.NoError(t, kin.Push([]byte(msg), msg))
}
