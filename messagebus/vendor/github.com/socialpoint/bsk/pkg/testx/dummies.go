package testx

// DummyCloser is useful for test doubles of io.Closers
// After the instance is closed, subsequent operations (different than calling Close again) must fail
type DummyCloser struct {
	isClosed bool
}

// IsClosed returns whether the instance is closed
func (c DummyCloser) IsClosed() bool {
	return c.isClosed
}

// Close always succeeds even if called multiple times
func (c *DummyCloser) Close() error {
	c.isClosed = true
	return nil
}
