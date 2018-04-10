package messenger

type dispatcher struct {
	driver DispatcherDriver
}

type DispatcherDriver interface {
	Dispatch(msg Message)
}

func NewDispatcher(driver DispatcherDriver) *dispatcher {
	return &dispatcher{driver: driver}
}

func (d *dispatcher) Dispatch(msg Message) {
	d.driver.Dispatch(msg)
}
