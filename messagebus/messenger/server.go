package messenger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Message interface {
	Topic() string
	Payload() map[string]interface{}
}

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

type Server interface {
	Serve()
}

type Consumer interface {
	Consume(topic string)
}

type httpServer struct {
}

func NewHttpServer() *httpServer {
	return &httpServer{}
}

type Handler interface {
	Handle(msg Message)
}

func (hs *httpServer) Serve() {
	http.HandleFunc("/", hs.handler)
	http.ListenAndServe(":80", nil)
}

func (hs *httpServer) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Wrong method")
		return
	}
	bodyReader, err := r.GetBody()
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return
	}

	message := &message{}
	err = json.Unmarshal(body, message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Wrong message")
		return
	}
}
