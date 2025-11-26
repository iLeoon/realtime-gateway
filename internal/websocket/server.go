package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/chatserver/pkg/logger"
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
		logger.Error("error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	client := &NewClient{conn: conn, send: make(chan []byte, 256)}
	logger.Info("A new client has been connected to the server")

	go client.readPump()
	go client.writePump()

}
