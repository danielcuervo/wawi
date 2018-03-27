package messenger

type Dispatcher interface {
	Dispatch(msg Message)
}

type kafkaDispatcher struct {
}

func NewKafkaDispatcher() *kafkaDispatcher {
	return &kafkaDispatcher{}
}

func (kd *kafkaDispatcher) Dispatch(msg Message) {
}
