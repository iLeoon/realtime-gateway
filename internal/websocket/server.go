package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/chatserver/pkg/logger"
)

type server struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
}

func newServer() *server {
	return &server{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Upgrading the http protocol into a websocket protocol
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func wsServer(s *server, w http.ResponseWriter, r *http.Request) {
	//the actual websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256), server: s}
	client.server.register <- client
	logger.Info("A new client has been connected to the server")
	go client.readPump()
	go client.writePump()

}
