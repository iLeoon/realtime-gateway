package websocket

import (
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

// WsServer manages all active WebSocket clients. It maintains a map of
// connected sessions and uses register/unregister channels to handle
// client lifecycle events
type wsServer struct {
	clients    map[uint32]*Client
	register   chan *Client
	unregister chan *Client
	SignalToWs chan uint32
}

// Create new websocket server
func NewWsServer() *wsServer {
	return &wsServer{
		clients:    make(map[uint32]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		SignalToWs: make(chan uint32),
	}
}

// Upgrading the http protocol into a websocket protocol
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// InitServerFunc upgrades the incoming HTTP request to a WebSocket
// connection, initializes a new client session using the provided Session
// implementation, registers the client, and starts the read and write pump
// goroutines for message handling.
func initServer(s *wsServer, w http.ResponseWriter, r *http.Request, tcpClient session.Session) {
	//the actual websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	client := &Client{
		conn:         conn,
		Send:         make(chan []byte, 256),
		server:       s,
		tcpClient:    tcpClient,
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
			err := client.tcpClient.OnConnect(client.ConnectionID)
			if err != nil {
				logger.Error("Error on encoding the connect packt", "Error", err)
				return
			}
		case client := <-s.unregister:
			if _, ok := s.clients[client.ConnectionID]; ok {
				//Remove the connectionID from the websocket map
				delete(s.clients, client.ConnectionID)
				//Remove the connectionID from the tcp server map
				err := client.tcpClient.DisConnect(client.ConnectionID)
				if err != nil {
					logger.Error("Error on encoding the connect packt", "Error", err)
					return
				}
				//Close the channel
				close(client.Send)
			}
		case id := <-s.SignalToWs:
			client := s.clients[id]
			delete(s.clients, client.ConnectionID)
			close(client.Send)
			client.conn.Close()

		}
	}
}

// Clients returns a pointer to the WebSocket server’s active client map.
// This allows external components—such as the Router—to access and route
// messages to currently connected WebSocket clients.
func (s *wsServer) Clients() *map[uint32]*Client {
	return &s.clients
}
