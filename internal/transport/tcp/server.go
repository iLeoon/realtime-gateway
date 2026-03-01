package tcp

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/protocol"
	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
	"github.com/iLeoon/realtime-gateway/internal/transport/tcp/worker"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TcpServer represents the central processing engine of the system. It is
// responsible for managing active WebSocket clients, receiving packets from
// the gateway, applying server-side logic, and routing messages to the
// appropriate clients.
type server struct {
	conf        *config.Config
	db          *pgxpool.Pool
	connections map[uint32]net.Conn // A map that holds connections to their raw tcp
	clients     map[uint32][]uint32 // A map that holds user ids to their connections
	ready       chan<- struct{}
	mu          sync.Mutex
	messagesCh  chan worker.Message
	done        chan struct{}
}

// Create a new instance of the TCP server.
func Start(c *config.Config, db *pgxpool.Pool, ready chan<- struct{}) {
	server := &server{
		conf:        c,
		db:          db,
		connections: make(map[uint32]net.Conn),
		clients:     make(map[uint32][]uint32),
		ready:       ready,
		done:        make(chan struct{}, 1),
		messagesCh:  make(chan worker.Message, 200),
	}

	worker.New(server.done, server.messagesCh, server.db)
	server.listen()
}

// Lanunches the server, this method must be invoked inside a separate
// goroutine because it blocks while listening for incoming packets.
func (s *server) listen() {
	const op errors.Op = "server.start"
	listner, err := net.Listen("tcp", s.conf.TcpPort)
	if err != nil {
		log.Error.Fatal("an error occured on creating tcp server", err)
		os.Exit(1)
	}

	log.Info.Println("TCP server is up and running")
	defer listner.Close()

	close(s.ready)

	// Listening to the connections
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Error.Println("an error occured while trying to connect a client", err)
			continue
		}
		go s.handleConn(conn)
		go s.pingReq(conn)
	}
}

// handleConn is the main packet dispatcher for the TcpServer. It receives
// decoded packets from the gateway and routes them to their corresponding
// handlers using a switch on the packet opcode.
//
// It uses type assertion to convert the generic BuildPayload
// its concrete SendMessagePacket type.
func (s *server) handleConn(conn net.Conn) {
	var userID uint32
	const op errors.Op = "server.handleConn"
	defer func() {
		log.Info.Println("The connection to the TCP server is terminated")
		conn.Close()
	}()

	for {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		// Call the decoder function on the connection to read
		// the incoming raw bytes and return the actual human-readable frame.
		frame, err := protocol.DecodeFrame(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info.Println("TCP connection is closed by peer(gateway)")
				return

			}
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Error.Println("unexpected tcp read error", err)
			return
		}

		// Route the frame to it's appropriate handler.
		// uses a type assertion to convert the generic BuildPayload interface into
		// its concrete *packet type
		switch p := frame.Payload.(type) {
		case *packets.ConnectPacket:
			userID = p.UserID
			if err := s.register(p, conn); err != nil {
				log.Error.Println("error on registering a new entery", err)
				return
			}
		case *packets.DisconnectPacket:
			s.unregister(p, conn)
		case *packets.SendMessagePacket:
			err := s.handleSendMessageReq(p, userID, conn)
			if err != nil {
				log.Error.Println("error on encoding response packet", err)
				return
			}
		case *packets.PongPacket:
			// We don't do anything we just renter the loop.
		default:
			log.Error.Printf("invalid packet type from gateway: %T", p)
			return
		}
		log.Info.Println("Decode packet", "packet", frame.Payload.String())
	}

}

func (s *server) writePacket(pkt packets.BuildPayload, conn net.Conn) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.writePacket"
	if err := conn.SetWriteDeadline(time.Now().Add(writeDuration)); err != nil {
		return errors.B(path, op, errors.Internal, fmt.Errorf("connection is unhealty: %v", err))
	}

	// Construct the frame, encode it, and then send it to the TCP server.
	frame := protocol.ConstructFrame(pkt)
	err := frame.EncodeFrame(conn)
	if err != nil {
		return errors.B(path, op, errors.Internal, err)
	}
	return nil
}

// handleSendMessage processes an inbound SendMessage packet.
func (s *server) handleSendMessageReq(pkt *packets.SendMessagePacket, userID uint32, conn net.Conn) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.handleSendMessageReq"

	type connInfo struct {
		connectionID  uint32
		rawConnection net.Conn
	}

	s.mu.Lock()
	connIDs, ok := s.clients[pkt.RecipientUserID]
	connections := make([]connInfo, 0, len(connIDs))
	if !ok {
		s.mu.Unlock()
		s.writePacket(&packets.ErrorPacket{Code: errors.NotFound, Message: "the connection entrey doesn't exist"}, conn)
		return errors.B(path, op, "couldn't find the connectionID within the map", errors.Internal)
	}

	// Fan-Out to all the user's connections
	for _, connID := range connIDs {
		if rawConn, ok := s.connections[connID]; ok {
			connections = append(connections, connInfo{connectionID: connID, rawConnection: rawConn})
		}
	}
	s.mu.Unlock()

	for _, v := range connections {
		resPkt := &packets.ResponseMessagePacket{
			ToConnectionID: v.connectionID,
			ResContent:     pkt.Content,
		}
		if err := s.writePacket(resPkt, v.rawConnection); err != nil {
			log.Error.Printf("operation %s: failed to write to this connection: %d: due to %v", op, v.connectionID, err)
		}
	}
	s.batchMessages(worker.Message{AuthorID: userID, ConversationID: pkt.ConversationID, Content: pkt.Content})

	return nil
}

func (s *server) pingReq(conn net.Conn) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.pingReq"
	ticker := time.NewTicker(pingTime)
	pkt := &packets.PingPacket{}
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.writePacket(pkt, conn); err != nil {
				return errors.B(path, op, errors.Internal, err)
			}
		case <-s.done:
			return nil
		}
	}
}

// registerConnectionIDs add a connecteionIDs and userIDs to their maps.
func (s *server) register(pkt *packets.ConnectPacket, conn net.Conn) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.register"
	s.mu.Lock()
	// To prevent overwriting existing connections.
	_, ok := s.connections[pkt.ConnectionID]
	if ok {
		s.mu.Unlock()
		s.writePacket(&packets.ErrorPacket{Code: errors.Client, Message: "connection already exists in the entrey"}, conn)
		return errors.B(path, op, errors.Client, fmt.Errorf("ConnectionID: %d already exists in the map", pkt.ConnectionID))
	}
	s.connections[pkt.ConnectionID] = conn
	s.clients[pkt.UserID] = append(s.clients[pkt.UserID], pkt.ConnectionID)
	s.mu.Unlock()
	return nil
}

// unRegisterConnectionIDs removes the connectionIDs and userIDs from the map
func (s *server) unregister(pkt *packets.DisconnectPacket, conn net.Conn) {
	conn.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, pkt.ConnectionID)
	clients := s.clients[pkt.UserID]
	// Create a new slice that doesn't share the same underlying array.
	filtered := make([]uint32, 0, len(clients))
	for _, client := range clients {
		if client != pkt.ConnectionID {
			filtered = append(filtered, client)
		}
	}
	if len(filtered) == 0 {
		delete(s.clients, pkt.UserID)
	} else {
		s.clients[pkt.UserID] = filtered
	}
}

func (s *server) batchMessages(message worker.Message) {
	const op errors.Op = "server.batchMessage"
	const path errors.PathName = "tcp/server"
	select {
	case s.messagesCh <- message:
	case <-s.done:
	default:
		log.Error.Println(errors.B(path, op, errors.ServiceUnavailable, "message dropped: worker buffer full"))
	}
}
