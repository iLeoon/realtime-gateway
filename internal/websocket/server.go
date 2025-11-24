package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//Upgrading the http protocol into a websocket protocol

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func wsServer(w http.ResponseWriter, r *http.Request) {
	//the actual websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("An Error occured while trying to upgrade connectin")
	}

	client := &Client{conn: conn, send: make(chan []byte, 256)}

	go client.readPump()

}
