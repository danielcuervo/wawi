package jwk_test

import (
	"errors"
	"testing"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/socialpoint/bsk/pkg/jwk"
	"github.com/stretchr/testify/assert"
)

var (
	someKey = &jwk.Key{}
	someErr = errors.New("oops")
	someTag = metrics.NewTag("foo", "bar")
)

func TestNewInstrumentReaderDecorator(t *testing.T) {
	assert := assert.New(t)

	recorder := metrics.NewRecorder()

	rf := jwk.ReaderFunc(func() ([]*jwk.Key, error) {
		return []*jwk.Key{someKey}, someErr
	})

	reader := jwk.InstrumentReader(rf, recorder, someTag)

	keys, err := reader.Read()

	assert.EqualError(err, someErr.Error())
	assert.Equal([]*jwk.Key{someKey}, keys)

	m := recorder.Get("jwk.reader.read_duration")
	mt, ok := m.(*metrics.RecorderTimer)
	assert.True(ok)
	assert.True(mt.StoppedTime.Sub(mt.StartedTime) > 0)
	assert.EqualValues(metrics.Tags{someTag}, mt.Tags())
}

func TestNewInstrumentWriterDecorator(t *testing.T) {
	assert := assert.New(t)

	recorder := metrics.NewRecorder()

	wf := jwk.WriterFunc(func(keys ...*jwk.Key) error {
		return someErr
	})

	writer := jwk.InstrumentWriter(wf, recorder, someTag)
	err := writer.Write(nil)

	assert.EqualError(err, someErr.Error())

	m := recorder.Get("jwk.writer.write_duration")
	mt, ok := m.(*metrics.RecorderTimer)
	assert.True(ok)
	assert.True(mt.StoppedTime.Sub(mt.StartedTime) > 0)
	assert.EqualValues(metrics.Tags{someTag}, mt.Tags())
}

func TestErrorCountingReaderDecorator(t *testing.T) {
	assert := assert.New(t)
	recorder := metrics.NewRecorder()

	stub := errors.New("error stub")

	fail := jwk.ReaderFunc(func() ([]*jwk.Key, error) { return nil, stub })

	_, err := jwk.InstrumentReader(fail, recorder).Read()
	assert.EqualValues(stub, err)

	m := recorder.Get("jwk.reader.read_error")
	counter, _ := m.(*metrics.RecorderCounter)
	assert.EqualValues(1, counter.Val())

	success := jwk.ReaderFunc(func() ([]*jwk.Key, error) { return nil, nil })

	_, err = jwk.InstrumentReader(success, recorder).Read()
	assert.NoError(err)

	m = recorder.Get("jwk.reader.read_error")
	counter, _ = m.(*metrics.RecorderCounter)
	assert.EqualValues(1, counter.Val())
}

func TestErrorCountingWriterDecorator(t *testing.T) {
	assert := assert.New(t)
	recorder := metrics.NewRecorder()

	stub := errors.New("error stub")

	fail := jwk.WriterFunc(func(keys ...*jwk.Key) error {
		return stub
	})

	err := jwk.InstrumentWriter(fail, recorder).Write()
	assert.EqualValues(stub, err)

	m := recorder.Get("jwk.writer.write_error")
	counter, _ := m.(*metrics.RecorderCounter)
	assert.EqualValues(1, counter.Val())

	success := jwk.WriterFunc(func(keys ...*jwk.Key) error { return nil })

	err = jwk.InstrumentWriter(success, recorder).Write()
	assert.NoError(err)

	m = recorder.Get("jwk.writer.write_error")
	counter, _ = m.(*metrics.RecorderCounter)
	assert.EqualValues(1, counter.Val())
}
