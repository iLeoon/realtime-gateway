package websocket

import (
	"container/list"
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
	clients     map[string][]*Client
	unregister  chan unregisterRequest
	signalToWs  chan ws.SignalToWsReq
	mu          sync.Mutex
	idleList    *list.List
	maxIdleTime time.Duration
	reaperCh    chan struct{}
}

// unregisterRequest is a custom struct that catches the reason for unregistering a client.
type unregisterRequest struct {
	client *Client
	reason string
}

// Create new websocket server
func NewWsServer() *wsServer {
	s := &wsServer{
		clients:    make(map[string][]*Client),
		unregister: make(chan unregisterRequest),
		signalToWs: make(chan ws.SignalToWsReq),
		idleList:   list.New(),
	}
	s.setMaxIdleTime(30 * time.Second)
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
		s.initWsServer(w, r, tcp)
	})
}

// initServerFunc upgrades the incoming HTTP request to a WebSocket
// connection, initializes a new client session using the provided Session
// implementation, registers the client, and starts the read and write pump
// goroutines for message handling.
func (s *wsServer) initWsServer(w http.ResponseWriter, r *http.Request, session session.InitiateSession) {
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
	s.mu.Lock()
	s.registerClient(client)
	s.putConn(client)
	s.mu.Unlock()

	go client.readPump()
	go client.writePump()
	go client.limiterFaucet()
}

func (s *wsServer) registerClient(client *Client) {
	//Add the connectionID to the websocket map
	s.clients[client.userID] = append(s.clients[client.userID], client)

	//Add the connectionID to the tcp server map
	err := client.tcpClient.OnConnect()
	if err != nil {
		logger.Error("Couldn't register this client", "ClientID", "Error", client.connectionID, err)
		delete(s.clients, client.userID)
	}
}

// register/unregister a connection to the clients map
// and signal to the ws when an error occures to disconnect the user.
func (s *wsServer) run() {
	for {
		select {
		case req := <-s.unregister:
			s.mu.Lock()
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
				err := req.client.tcpClient.OnDisConnect()
				if err != nil {
					logger.Error("Couldn't unregister this client", "ClientID", req.client.connectionID, "Error", err)
				}

				// permanently close the connection.
				req.client.Terminate()
			}
			s.mu.Unlock()

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

func (s *wsServer) removeConnections(clients []*Client, target uint32) []*Client {
	filtered := clients[:0]
	for _, c := range clients {
		if c.connectionID != target {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (s *wsServer) GetClient(userID string, connectionID uint32) (ws.WsClient, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clients, ok := s.clients[userID]
	if !ok {
		return nil, false
	}

	for _, client := range clients {
		if client.connectionID == connectionID {
			return client, true
		}
	}
	return nil, false
}

func (s *wsServer) SignalToWs(req ws.SignalToWsReq) {
	s.signalToWs <- req
}

func (s *wsServer) putConn(client *Client) {
	client.isActive = false
	client.lastActiveAt = time.Now()

	s.putConnLocked(client)
}

func (s *wsServer) putConnLocked(client *Client) {
	// Check for duplicates.
	if client.idleElement != nil {
		s.idleList.Remove(client.idleElement)
		// client.idleElement = nil
	}

	// Only add the inactive clients to the list.
	if !client.isActive && len(s.clients) != 0 {
		client.idleElement = s.idleList.PushBack(client)
	}

	if s.reaperCh == nil && s.maxIdleTime > 0 {
		s.reaperCh = make(chan struct{}, 1)
		go s.startReaper(s.maxIdleTime)
	}

}

func (s *wsServer) reclaimConnLocked(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client.idleElement == nil {
		client.isActive = true
		return
	}

	s.idleList.Remove(client.idleElement)
	client.idleElement = nil
}

func (s *wsServer) setMaxIdleTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d < 0 {
		d = 0
	}

	// If we decided to configure the time just wake the reaper with the new settings.
	if d > 0 && d < s.maxIdleTime && s.reaperCh != nil {
		select {
		case s.reaperCh <- struct{}{}:
		default:
		}
	}
	s.maxIdleTime = d

	if s.reaperCh == nil && s.maxIdleTime > 0 && len(s.clients) != 0 {
		s.reaperCh = make(chan struct{}, 1)
		go s.startReaper(d)
	}

}

func (s *wsServer) startReaper(d time.Duration) {
	const minInterval = time.Second

	if d < minInterval {
		d = minInterval
	}

	t := time.NewTimer(d)

	for {
		select {
		case <-t.C:
		case <-s.reaperCh:
		}

		s.mu.Lock()
		if s.maxIdleTime <= 0 || len(s.clients) == 0 {
			s.reaperCh = nil
			s.mu.Unlock()
			return
		}

		d, clients := s.cleanReaperLocked(d)
		s.mu.Unlock()

		if clients != nil {
			for _, c := range clients {
				s.unregister <- unregisterRequest{client: c, reason: "The connection has been idle for too long"}
			}
		}

		if d < minInterval {
			d = minInterval
		}

		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		t.Reset(d)
	}

}

func (s *wsServer) cleanReaperLocked(d time.Duration) (time.Duration, []*Client) {
	if s.maxIdleTime <= 0 || s.idleList.Len() == 0 {
		s.reaperCh = nil
		return d, nil
	}

	var closing []*Client
	idleSince := time.Now().Add(-s.maxIdleTime)

	var next *list.Element
	for e := s.idleList.Front(); e != nil; e = next {
		next = e.Next()
		client := e.Value.(*Client)
		if client.lastActiveAt.Before(idleSince) {
			closing = append(closing, client)
			s.idleList.Remove(e)
			client.idleElement = nil
		} else {
			d = client.lastActiveAt.Sub(idleSince)

			break

		}
	}
	return d, closing
}
