package testx

// CallCounter counts home many times a method call has been made
type CallCounter struct {
	numCalls int
}

// NumCalls returns how many times the Main Method has been called (successfully or not)
func (c CallCounter) NumCalls() int {
	return c.numCalls
}

// IncCalls increments the number of calls
func (c *CallCounter) IncCalls() {
	c.numCalls++
}
