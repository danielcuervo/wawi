package driver

type message struct {
	topic   string
	payload map[string]interface{}
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Payload() map[string]interface{} {
	return m.payload
}
