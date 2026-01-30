package websocket

import (
	"container/list"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

type signalToWsReq struct {
	userId       string
	connectionId uint32
}

// server manages all active WebSocket clients. It maintains a map of
// connected sessions and uses register/unregister channels to handle
// client lifecycle events
type server struct {
	clients     map[string][]*client
	config      *config.Config
	unregister  chan unregisterRequest
	signalToWs  chan signalToWsReq
	mu          sync.Mutex
	idleList    *list.List
	maxIdleTime time.Duration
	reaperCh    chan struct{}
}

// unregisterRequest is a custom struct that catches the reason for unregistering a client.
type unregisterRequest struct {
	client *client
	reason string
}

// Create new websocket server
func New(config *config.Config) *server {
	s := &server{
		clients:    make(map[string][]*client),
		unregister: make(chan unregisterRequest),
		signalToWs: make(chan signalToWsReq),
		idleList:   list.New(),
		config:     config,
	}
	s.setMaxIdleTime(30 * time.Second)
	go s.run()
	return s
}

// Handle constructs and returns an http handler responsible for handling
// webSocket requests then pass it to start
func (s *server) Handle(tcp session.InitiateSession) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Start(w, r, tcp)
	})
}

// acceptWs upgrades the incoming HTTP request to a WebSocket
// connection, initializes a new client session using the provided Session
// implementation, registers the client, and starts the read and write pump
// goroutines for message handling.
func (s *server) Start(w http.ResponseWriter, r *http.Request, session session.InitiateSession) {
	// the upgrader configuration
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return s.config.FrontEndOrigin == r.Header.Get("Origin")
		},
	}
	connectionID := rand.Uint32()
	userId, ok := ctx.UserId(r.Context())

	if !ok {
		logger.Error("Couldn't extract the ID from the request")
		return
	}

	// upgrade the websocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Error on upgrading raw tcp connection into websocekt", "Error", err)
		return
	}

	tcpClient, err := session.NewClient(userId, connectionID)
	if err != nil {
		logger.Error("Error on initializing a new tcp client for the connection between websocket and tcp server")
		return
	}

	client := &client{
		userId:        userId,
		conn:          conn,
		send:          make(chan []byte, 256),
		server:        s,
		tcpClient:     tcpClient,
		burstyLimiter: make(chan time.Time, 3),
		connectionId:  connectionID,
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

func (s *server) Send(userId string, connectionId uint32, message []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	clients, ok := s.clients[userId]
	if !ok {
		return fmt.Errorf("no client was found with that userId: %d", userId)
	}

	for i := range clients {
		if clients[i].connectionId == connectionId {
			client := clients[i]
			client.enqueue(message)
			return nil
		}
	}
	return fmt.Errorf("no clients were found with that connectionId: %d", connectionId)
}

func (s *server) registerClient(client *client) {
	//Add the connectionID to the websocket map
	s.clients[client.userId] = append(s.clients[client.userId], client)

	//Add the connectionID to the tcp server map
	err := client.tcpClient.OnConnect()
	if err != nil {
		logger.Error("Couldn't register this client", "ClientID", "Error", client.connectionId, err)
		delete(s.clients, client.userId)
	}
}

// Unregister unregisters a connection to the clients map
// and signal to the ws when an error occures to disconnect the user.
func (s *server) run() {
	for {
		select {
		case req := <-s.unregister:
			s.mu.Lock()
			if clients, ok := s.clients[req.client.userId]; ok {
				logger.Info("Terminating the connection", "ID", req.client.connectionId, "Reason", req.reason)

				//Remove the connectionID from the websocket map
				clients = s.removeConnections(clients, req.client.connectionId)
				if len(clients) == 0 {
					delete(s.clients, req.client.userId)
				} else {
					s.clients[req.client.userId] = clients
				}

				//Remove the connectionID from the tcp server map
				err := req.client.tcpClient.OnDisConnect()
				if err != nil {
					logger.Error("Couldn't unregister this client", "ClientID", req.client.connectionId, "Error", err)
				}

				// permanently close the connection.
				req.client.Terminate()
			}
			s.mu.Unlock()

		case signal := <-s.signalToWs:
			if clients, ok := s.clients[signal.userId]; ok {
				logger.Info("Signal received to kill connection", "ID", signal.connectionId, "UserID", signal.userId)

				updatedClients := s.removeConnections(clients, signal.connectionId)
				if len(updatedClients) == 0 {
					delete(s.clients, signal.userId)
				} else {
					s.clients[signal.userId] = updatedClients
				}

				for i := range clients {
					if clients[i].connectionId == signal.connectionId {
						clients[i].Terminate()
						break
					}
				}
			}
		}
	}
}

func (s *server) removeConnections(clients []*client, target uint32) []*client {
	filtered := clients[:0]
	for i := range clients {
		if clients[i].connectionId != target {
			filtered = append(filtered, clients[i])
		}
	}
	return filtered
}

func (s *server) Signal(userId string, connectionId uint32) {
	s.signalToWs <- signalToWsReq{
		userId:       userId,
		connectionId: connectionId,
	}
}

func (s *server) putConn(client *client) {
	client.isActive = false
	client.lastActiveAt = time.Now()

	s.putConnLocked(client)
}

func (s *server) putConnLocked(client *client) {
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

func (s *server) reclaimConnLocked(client *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client.idleElement == nil {
		client.isActive = true
		return
	}

	s.idleList.Remove(client.idleElement)
	client.idleElement = nil
}

func (s *server) setMaxIdleTime(d time.Duration) {
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

func (s *server) startReaper(d time.Duration) {
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
			for i := range clients {
				s.unregister <- unregisterRequest{client: clients[i], reason: "The connection has been idle for too long"}
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

func (s *server) cleanReaperLocked(d time.Duration) (time.Duration, []*client) {
	if s.maxIdleTime <= 0 || s.idleList.Len() == 0 {
		s.reaperCh = nil
		return d, nil
	}

	var closing []*client
	idleSince := time.Now().Add(-s.maxIdleTime)

	var next *list.Element
	for e := s.idleList.Front(); e != nil; e = next {
		next = e.Next()
		client := e.Value.(*client)
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
