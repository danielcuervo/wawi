package main

import (
	"github.com/danielcuervo/wawi/messenger/messenger"
)

func main() {
	messengerServer := messenger.NewHttpServer()
	messengerServer.Serve()
}
