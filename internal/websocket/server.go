package websocket

import (
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/session"
)

type wsServer struct {
	clients    map[*Client]uint32
	register   chan *Client
	unregister chan *Client
}

func newWsServer() *wsServer {
	return &wsServer{
		clients:    make(map[*Client]uint32),
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

func initServer(s *wsServer, w http.ResponseWriter, r *http.Request, tcp session.Session) {
	//the actual websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	client := &Client{
		conn:         conn,
		send:         make(chan []byte, 256),
		server:       s,
		transporter:  tcp,
		connectionID: rand.Uint32(),
	}

	client.server.register <- client
	logger.Info("A new client has been connected to the server")
	go client.readPump()
	go client.writePump()

}
