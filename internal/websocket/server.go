package websocket

import (
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/session"
)

type wsServer struct {
	clients    map[uint32]*Client
	register   chan *Client
	unregister chan *Client
}

func NewWsServer() *wsServer {
	return &wsServer{
		clients:    make(map[uint32]*Client),
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
		Send:         make(chan []byte, 256),
		server:       s,
		transporter:  tcp,
		ConnectionID: rand.Uint32(),
	}

	client.server.register <- client
	logger.Info("A new client has been connected to the server")
	go client.readPump()
	go client.writePump()

}

func (s *wsServer) run() {
	for {
		select {
		case client := <-s.register:
			//Add the connectionID to the websocket map
			s.clients[client.ConnectionID] = client
			//Add the connectionID to the tcp server map
			client.transporter.OnConnect(client.ConnectionID)
		case client := <-s.unregister:
			if _, ok := s.clients[client.ConnectionID]; ok {
				//Remove the connectionID from the websocket map
				delete(s.clients, client.ConnectionID)
				//Remove the connectionID from the tcp server map
				client.transporter.DisConnect(client.ConnectionID)
				//Close the channel
				close(client.Send)
			}

		}
	}
}

func (s *wsServer) Clients() *map[uint32]*Client {
	return &s.clients
}
