package tcp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/protocol"
	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

// TcpServer represents the central processing engine of the system. It is
// responsible for managing active WebSocket clients, receiving packets from
// the gateway, applying server-side logic, and routing messages to the
// appropriate clients.
type server struct {
	conf *config.Config
	// A map of connected client IDs to their active WebSocket
	// sessions, allowing direct message delivery.
	clients map[uint32]net.Conn
	ready   chan<- struct{}
	mu      sync.Mutex
}

// Create a new instance of the TCP server.
func NewServer(conf *config.Config, ready chan<- struct{}) *server {
	server := &server{
		conf:    conf,
		clients: make(map[uint32]net.Conn),
		ready:   ready,
	}
	server.start()
	return server
}

// Lanunches the server, this method must be invoked inside a separate
// goroutine because it blocks while listening for incoming packets.
func (s *server) start() {
	listner, err := net.Listen("tcp", s.conf.TcpPort)
	if err != nil {
		log.Error.Fatal("An error occured on creating tcp server", "Error", err)
		os.Exit(1)
	}

	log.Info.Println("TCP server is up and running")
	defer listner.Close()

	close(s.ready)

	// Listening to the connections
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Error.Println("An error occured while trying to connect a client", "Error", err)
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

			log.Error.Println("Unexpected tcp read error", "error", err)
			return
		}

		// Route the frame to it's appropriate handler.
		// uses a type assertion to convert the generic BuildPayload interface into
		// its concrete *packet type
		switch p := frame.Payload.(type) {
		case *packets.ConnectPacket:
			s.registerConnectionIDs(p, conn)
		case *packets.DisconnectPacket:
			s.unregisterConnectionIDs(p, conn)
		case *packets.SendMessagePacket:
			err := s.handleSendMessageReq(p)
			if err != nil {
				log.Error.Println("Error on encoding response packet", "Error", err)
				return
			}
		case *packets.PongPacket:
			// We don't do anything we just renter the loop.
		default:
			log.Error.Println("Invalid packet type from gateway: %T", p)
			return
		}
		log.Info.Println("Decode packet", "packet", frame.Payload.String())
	}

}

func (s *server) writePacket(pkt packets.BuildPayload, conn net.Conn) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeDuration)); err != nil {
		return fmt.Errorf("connection is unhealty: %w", err)
	}

	// Construct the frame, encode it, and then send it to the TCP server.
	frame := protocol.ConstructFrame(pkt)
	err := frame.EncodeFrame(conn)
	if err != nil {
		return err
	}

	return nil
}

// handleSendMessage processes an inbound SendMessage packet.
func (s *server) handleSendMessageReq(pkt *packets.SendMessagePacket) error {
	var recipient uint32
	for id, _ := range s.clients {
		if id != pkt.ConnectionID {
			recipient = id
		}
	}

	resPkt := &packets.ResponseMessagePacket{
		ToConnectionID: recipient,
		ResContent:     pkt.Content,
	}

	recipientConnection, ok := s.clients[recipient]
	if !ok {
		return fmt.Errorf("couldn't find the connectionID within the map")
	}

	err := s.writePacket(resPkt, recipientConnection)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) pingReq(conn net.Conn) error {
	ticker := time.NewTicker(pingTime)
	pkt := &packets.PingPacket{}
	defer ticker.Stop()

	for range ticker.C {
		err := s.writePacket(pkt, conn)
		if err != nil {
			return err
		}

	}
	return nil
}

// registerConnectionIDs add a connecteionID to the map associated with it's tcp connection.
func (s *server) registerConnectionIDs(pkt *packets.ConnectPacket, conn net.Conn) {
	s.mu.Lock()
	s.clients[pkt.ConnectionID] = conn
	s.mu.Unlock()
}

// unRegisterConnectionIDs removes the connectionID from the map
func (s *server) unregisterConnectionIDs(pkt *packets.DisconnectPacket, conn net.Conn) {
	s.mu.Lock()
	conn.Close()
	delete(s.clients, pkt.ConnectionID)
	s.mu.Unlock()
}
