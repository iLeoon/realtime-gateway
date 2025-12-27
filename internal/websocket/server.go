package websocket

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/session"
	"github.com/iLeoon/realtime-gateway/pkg/ws"
)

// WsServer manages all active WebSocket clients. It maintains a map of
// connected sessions and uses register/unregister channels to handle
// client lifecycle events
type wsServer struct {
	clients    map[uint32]*Client
	register   chan *Client
	unregister chan *Client
	signalToWs chan uint32
	mu         sync.Mutex
}

// Create new websocket server
func NewWsServer() *wsServer {
	s := &wsServer{
		clients:    make(map[uint32]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		signalToWs: make(chan uint32),
	}
	go s.run()
	return s

}

// Upgrading the http protocol into a websocket protocol
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Start constructs and returns an http.Handler responsible for handling
// WebSocket upgrade requests. It upgrades incoming HTTP requests,
// creates WebSocket clients, and registers them with the gateway.
func (s *wsServer) Start(conf *config.Config, tcp session.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.initServer(w, r, tcp)
	})
}

// initServerFunc upgrades the incoming HTTP request to a WebSocket
// connection, initializes a new client session using the provided Session
// implementation, registers the client, and starts the read and write pump
// goroutines for message handling.
func (s *wsServer) initServer(w http.ResponseWriter, r *http.Request, tcpClient session.Session) {

	userID, ok := ctx.GetUserIDCtx(r.Context())
	fmt.Println(ok)
	fmt.Println(userID)

	//the actual websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	client := &Client{
		conn:         conn,
		send:         make(chan []byte, 256),
		server:       s,
		tcpClient:    tcpClient,
		connectionID: rand.Uint32(),
	}
	s.register <- client
	logger.Info("A new client has been connected to the server")

	go client.readPump()
	go client.writePump()

}

// run register/unregister a connection to the clients map
// and signal to the ws when an error occures to disconnect the user.
func (s *wsServer) run() {
	for {
		select {
		case client := <-s.register:
			//Add the connectionID to the websocket map
			s.clients[client.connectionID] = client
			//Add the connectionID to the tcp server map
			err := client.tcpClient.OnConnect(client.connectionID)
			if err != nil {
				logger.Error("Error on encoding the connect packt", "Error", err)
				return
			}
		case client := <-s.unregister:
			if _, ok := s.clients[client.connectionID]; ok {
				//Remove the connectionID from the websocket map
				delete(s.clients, client.connectionID)
				//Remove the connectionID from the tcp server map
				err := client.tcpClient.DisConnect(client.connectionID)
				if err != nil {
					logger.Error("Error on encoding the connect packt", "Error", err)
					return
				}
				//Close the channel
				close(client.send)
			}
		case id := <-s.signalToWs:
			client := s.clients[id]
			delete(s.clients, client.connectionID)
			close(client.send)
			client.conn.Close()

		}
	}
}

func (s *wsServer) GetClient(connectionID uint32) ws.WsClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clients[connectionID]
}

func (s *wsServer) SignalToWs(connectionID uint32) {
	s.signalToWs <- connectionID
}
