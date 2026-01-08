package websocket

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

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
	clients    map[string][]*Client
	register   chan *Client
	unregister chan unregisterRequest
	signalToWs chan ws.SignalToWsReq
	mu         sync.Mutex
}

type unregisterRequest struct {
	client *Client
	reason string
}

// Create new websocket server
func NewWsServer() *wsServer {
	s := &wsServer{
		clients:    make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan unregisterRequest),
		signalToWs: make(chan ws.SignalToWsReq),
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
func (s *wsServer) Start(conf *config.Config, tcp session.InitiateSession) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.initServer(w, r, tcp)
	})
}

// initServerFunc upgrades the incoming HTTP request to a WebSocket
// connection, initializes a new client session using the provided Session
// implementation, registers the client, and starts the read and write pump
// goroutines for message handling.
func (s *wsServer) initServer(w http.ResponseWriter, r *http.Request, session session.InitiateSession) {

	connectionID := rand.Uint32()
	userID, ok := ctx.GetUserIDCtx(r.Context())
	if !ok {
		logger.Error("Couldn't extract the ID from the request")
		return
	}

	//the actual websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	tcpClient, err := session.NewTCPClient(userID, connectionID)
	if err != nil {
		logger.Error("Error on initializing a new tcp client for the connection between websocket and tcp server")
		return
	}

	client := &Client{
		userID:        userID,
		conn:          conn,
		send:          make(chan []byte, 256),
		server:        s,
		tcpClient:     tcpClient,
		burstyLimiter: make(chan time.Time, 3),
		connectionID:  connectionID,
		done:          make(chan struct{}),
	}
	s.register <- client

	go client.readPump()
	go client.writePump()
	go client.limiterFaucet()

}

// register/unregister a connection to the clients map
// and signal to the ws when an error occures to disconnect the user.
func (s *wsServer) run() {
	for {
		select {
		case client := <-s.register:
			//Add the connectionID to the websocket map
			s.clients[client.userID] = append(s.clients[client.userID], client)
			//Add the connectionID to the tcp server map
			err := client.tcpClient.OnConnect(client.connectionID)
			if err != nil {
				logger.Error("Couldn't register this client", "ClientID", "Error", client.connectionID, err)
				delete(s.clients, client.userID)
				continue
			}

		case req := <-s.unregister:
			if clients, ok := s.clients[req.client.userID]; ok {
				logger.Info("Terminating the connection", "ID", req.client.connectionID, "Reason", req.reason)

				//Remove the connectionID from the websocket map
				clients = s.removeConnections(clients, req.client.connectionID)
				if len(clients) == 0 {
					delete(s.clients, req.client.userID)
				} else {
					s.clients[req.client.userID] = clients
				}

				//Remove the connectionID from the tcp server map
				err := req.client.tcpClient.DisConnect(req.client.connectionID)
				if err != nil {
					logger.Error("Couldn't unregister this client", "ClientID", "Error", req.client.connectionID, err)
				}

				// permanently close the connection.
				req.client.Terminate()
			}
		case signal := <-s.signalToWs:
			if clients, ok := s.clients[signal.UserID]; ok {
				logger.Info("Signal received to kill connection", "ID", signal.ConnectionID, "UserID", signal.UserID)

				updatedClients := s.removeConnections(clients, signal.ConnectionID)
				if len(updatedClients) == 0 {
					delete(s.clients, signal.UserID)
				} else {
					s.clients[signal.UserID] = updatedClients
				}

				for _, c := range clients {
					if c.connectionID == signal.ConnectionID {
						c.Terminate()
						break
					}
				}
			}
		}
	}
}

// func (s *wsServer) GetClient(userID string) (ws.WsClient, bool) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	client, ok := s.clients[userID]
// 	return client, ok
// }

func (s *wsServer) SignalToWs(req ws.SignalToWsReq) {
	s.signalToWs <- req
}

func (s *wsServer) removeConnections(clients []*Client, target uint32) []*Client {
	filtered := clients[:0]
	for _, c := range clients {
		if c.connectionID != target {
			filtered = append(filtered, c)
		}
	}

	return filtered

}
