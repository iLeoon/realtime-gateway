package websocket

import (
	"log"
	"net/http"

	"github.com/iLeoon/chatserver/internal/config"
)

func Start() {
	conf := config.Load()

	http.HandleFunc("/ws", wsServer)
	log.Println("The websocekt server is up and running..")

	log.Fatal(http.ListenAndServe(conf.Websocket.Port, nil))
}
