package main

import (
	"github.com/danielcuervo/wawi/messagebus/messenger"
)

func main() {
	messengerServer := messenger.NewHttpServer()
	messengerServer.Serve()
}
