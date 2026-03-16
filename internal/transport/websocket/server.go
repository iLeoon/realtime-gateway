package websocket

import (
	"container/list"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

const path errors.PathName = "websocket/server"

type Client interface {
	Enqueue(message []byte)
	Terminate(code int, reason string, op errors.Op)
	ConnectionID() uint32
}

type signalToWsReq struct {
	userID       string
	connectionID uint32
	code         int
	reason       string
}

// server manages all active WebSocket clients. It maintains a map of
// connected sessions and uses register/unregister channels to handle
// client lifecycle events
type server struct {
	clients     map[string][]Client
	c           *config.Config
	signalToWs  chan signalToWsReq
	mu          sync.Mutex
	// idleList to identify the idle connections in the websocket server and disconnect them.
	idleList    *list.List
	maxIdleTime time.Duration
	reaperCh    chan struct{}
}

// New create a new websocket server instance
func New(c *config.Config) *server {
	s := &server{
		clients:    make(map[string][]Client),
		signalToWs: make(chan signalToWsReq, 5),
		idleList:   list.New(),
		c:          c,
	}
	s.setMaxIdleTime(5 * time.Minute)
	go s.run()
	return s
}

// Handle constructs and returns an http handler responsible for handling
// webSocket requests then pass it to start
func (s *server) Handle(tcp session.InitiateSession) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		s.Start(w, r, tcp)
	}
}

// Start upgrades the incoming HTTP request to a WebSocket
// connection, initializes a new client session using the provided Session
// implementation, registers the client, and starts the read and write pump
// goroutines for message handling.
func (s *server) Start(w http.ResponseWriter, r *http.Request, session session.InitiateSession) {
	// the upgrader configuration
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return s.c.Cors == origin
		},
	}
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Error.Println(path, errors.Op("server.start"), err)
		return
	}
	connectionID := binary.LittleEndian.Uint32(b[:])

	userID, ok := ctx.UserID(r.Context())
	if !ok {
		log.Error.Println("couldn't extract the ID from the request")
		return
	}

	// upgrade the websocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error.Println("error on upgrading raw tcp connection into websocekt", err)
		return
	}

	tcpClient, err := session.NewClient(userID, connectionID)
	if err != nil {
		log.Error.Println("error on initializing a new tcp client for the connection between websocket and tcp server")
		conn.Close()
		return
	}

	client := &client{
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
	// Check if the connection was successfully registred to the map
	registered := s.registerClient(client)
	s.mu.Unlock()

	if registered {
		s.putConn(client)
	} else {
		conn.Close()
		return
	}

	go client.readPump()
	go client.writePump()
	go client.limiterFaucet()
}

func (s *server) Send(userID string, connectionID uint32, message []byte) error {
	s.mu.Lock()
	clients, ok := s.clients[userID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("no client was found with that userID: %s", userID)
	}
	var target Client
	for i := range clients {
		if clients[i].ConnectionID() == connectionID {
			target = clients[i]
			break
		}
	}

	s.mu.Unlock()
	if target == nil {
		return fmt.Errorf("no clients were found with that connectionID: %d", connectionID)
	}
	target.Enqueue(message)
	return nil
}

func (s *server) registerClient(c *client) bool {
	//Add the connectionID to the websocket map
	s.clients[c.userID] = append(s.clients[c.userID], c)

	//Add the connectionID to the tcp server map
	err := c.tcpClient.OnConnect()
	if err != nil {
		log.Error.Printf("couldn't register this client: %d due to :%v", c.connectionID, err)
		clients := s.removeConnections(s.clients[c.userID], c.connectionID)
		if len(clients) == 0 {
			delete(s.clients, c.userID)
		} else {
			s.clients[c.userID] = clients
		}
		return false
	}
	return true
}

// run handles signals from the TCP layer to terminate specific connections.
func (s *server) run() {
	const op errors.Op = "server.run"
	for signal := range s.signalToWs {
		s.mu.Lock()
		clients, ok := s.clients[signal.userID]
		var target Client
		if ok {
			for i := range clients {
				if clients[i].ConnectionID() == signal.connectionID {
					target = clients[i]
					break
				}
			}
		}
		s.mu.Unlock()
		if target != nil {
			log.Info.Println("Signal received to kill connection", "ID", signal.connectionID, "UserID", signal.userID)
			target.Terminate(signal.code, signal.reason, op)
		}

	}
}

func (s *server) removeConnections(clients []Client, target uint32) []Client {
	filtered := make([]Client, 0, len(clients))
	for i := range clients {
		if clients[i].ConnectionID() != target {
			filtered = append(filtered, clients[i])
		}
	}
	return filtered
}

func (s *server) removeClient(c *client) {
	s.mu.Lock()
	if c.idleElement != nil {
		s.idleList.Remove(c.idleElement)
		c.idleElement = nil
	}
	clients := s.removeConnections(s.clients[c.userID], c.connectionID)
	if len(clients) == 0 {
		delete(s.clients, c.userID)
	} else {
		s.clients[c.userID] = clients
	}
	s.mu.Unlock()
}

func (s *server) Signal(userID string, connectionID uint32, code int, reason string) {
	const op errors.Op = "server.Signal"
	select {
	case s.signalToWs <- signalToWsReq{userID: userID, connectionID: connectionID, code: code, reason: reason}:
	default:
		err := errors.B(path, op)
		log.Error.Println("the channel buffer is full", err)
		return
	}
}

func (s *server) putConn(client *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	client.isActive = false
	client.lastActiveAt = time.Now()
	s.putConnLocked(client)
}

func (s *server) putConnLocked(client *client) {
	// Check for duplicates.
	if client.idleElement != nil {
		s.idleList.Remove(client.idleElement)
		client.idleElement = nil
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

func (s *server) reclaimConn(client *client) {
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
	const op errors.Op = "server.startReaper"
	var clients []*client
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

		d, clients = s.cleanReaperLocked(d)
		s.mu.Unlock()

		for i := range clients {
			clients[i].Terminate(websocket.CloseGoingAway, "idle for too long", op)
			log.Info.Printf("connection %d (user %s) has been idle for too long", clients[i].ConnectionID(), clients[i].userID)
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
	const op errors.Op = "server.cleanReaperLocked"
	if s.maxIdleTime <= 0 || s.idleList.Len() == 0 {
		// Nothing to evict — sleep a full interval. Do NOT touch reaperCh;
		// the goroutine is still running and owns its own lifecycle.
		return s.maxIdleTime, nil
	}

	var closing []*client
	idleSince := time.Now().Add(-s.maxIdleTime)

	var next *list.Element
	for e := s.idleList.Front(); e != nil; e = next {
		next = e.Next()
		client, ok := e.Value.(*client)
		if !ok {
			wrapErr := errors.B(path, op, fmt.Errorf("expected *client in the list, but found %v", client))
			log.Error.Println(wrapErr)
			continue
		}

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
