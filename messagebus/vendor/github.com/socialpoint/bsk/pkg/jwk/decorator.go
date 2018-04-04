package jwk

import "github.com/socialpoint-labs/bsk/metrics"

// InstrumentReader returns reader decorated with instrumentation capabilities
func InstrumentReader(reader Reader, m metrics.Metrics, tags ...metrics.Tag) Reader {
	return ReaderFunc(func() ([]*Key, error) {
		t := m.Timer("jwk.reader.read_duration", tags...)

		t.Start()
		defer t.Stop()

		keys, err := reader.Read()

		if err != nil {
			m.Counter("jwk.reader.read_error", tags...).Inc()
		}

		return keys, err
	})
}

// InstrumentWriter returns a writer decorated with instrumentation capabilities
func InstrumentWriter(writer Writer, m metrics.Metrics, tags ...metrics.Tag) Writer {
	return WriterFunc(func(keys ...*Key) error {
		t := m.Timer("jwk.writer.write_duration", tags...)

		t.Start()
		defer t.Stop()

		err := writer.Write(keys...)

		if err != nil {
			m.Counter("jwk.writer.write_error", tags...).Inc()
		}

		return err
	})
}
