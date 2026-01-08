package tcp

import (
	"net"
	"os"
	"sync"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/protocol"
	"github.com/iLeoon/realtime-gateway/pkg/protocol/packets"
)

// TcpServer represents the central processing engine of the system. It is
// responsible for managing active WebSocket clients, receiving packets from
// the gateway, applying server-side logic, and routing messages to the
// appropriate clients.
type tcpServer struct {
	conf *config.Config
	// A map of connected client IDs to their active WebSocket
	// sessions, allowing direct message delivery.
	clients map[uint32]net.Conn
	ready   chan<- struct{}
	mu      sync.Mutex
}

// Create a new instance of the TCP server.
func NewTcpServer(conf *config.Config, ready chan<- struct{}) *tcpServer {
	server := &tcpServer{
		conf:    conf,
		clients: make(map[uint32]net.Conn),
		ready:   ready,
	}
	server.start()
	return server

}

// Lanunches the server, this method must be invoked inside a separate
// goroutine because it blocks while listening for incoming packets.
func (t *tcpServer) start() {
	listner, err := net.Listen("tcp", t.conf.TcpPort)
	if err != nil {
		logger.Error("An error occured on creating tcp server", "Error", err)
		os.Exit(1)
	}

	logger.Info("TCP server is up and running")
	defer listner.Close()

	close(t.ready)

	// Listening to the connections
	for {
		conn, err := listner.Accept()
		if err != nil {
			logger.Error("An error occured while trying to connect a client", "Error", err)
			continue
		}
		go t.handleConn(conn)
	}
}

// handleConn is the main packet dispatcher for the TcpServer. It receives
// decoded packets from the gateway and routes them to their corresponding
// handlers using a switch on the packet opcode.
//
// It uses type assertion to convert the generic BuildPayload
// its concrete SendMessagePacket type.
func (t *tcpServer) handleConn(conn net.Conn) {
	defer func() {
		logger.Info("Tcp server connection is terminated")
		conn.Close()
	}()

	for {

		// Call the decoder function on the connection to read
		// the incoming raw bytes and return the actual human-readable frame.
		frame, err := protocol.DecodeFrame(conn)
		if err != nil {
			logger.Error("Invalid data from gateway", "Error", err)
			return
		}

		// Route the frame to it's appropriate handler.
		// uses a type assertion to convert the generic BuildPayload interface into
		// its concrete *packet type
		switch p := frame.Payload.(type) {
		case *packets.ConnectPacket:
			t.registerConnectionIDs(p, conn)
		case *packets.DisconnectPacket:
			t.unregisterConnectionIDs(p, conn)
		case *packets.SendMessagePacket:
			err := t.handleSendMessageReq(p)
			if err != nil {
				logger.Error("Error on encoding response packet", "Error", err)
				return
			}
		default:
			logger.Error("Invalid packet type from gateway: %T", p)
			return
		}
		logger.Debug("Decode packet", "packet", frame.Payload.String())
	}

}

// handleSendMessage processes an inbound SendMessage packet.
func (t *tcpServer) handleSendMessageReq(pkt *packets.SendMessagePacket) error {
	var recipient uint32
	for id, _ := range t.clients {
		if id != pkt.ConnectionID {
			recipient = id
		}
	}

	resPkt := &packets.ResponseMessagePacket{
		ToConnectionID: recipient,
		ResContent:     pkt.Content,
	}

	frame := protocol.ConstructFrame(resPkt)
	err := frame.EncodeFrame(t.clients[recipient])
	if err != nil {
		return err
	}
	return nil
}

// registerConnectionIDs add a connecteionID to the map associated with it's tcp connection.
func (t *tcpServer) registerConnectionIDs(pkt *packets.ConnectPacket, conn net.Conn) {
	t.mu.Lock()
	t.clients[pkt.ConnectionID] = conn
	t.mu.Unlock()
}

// unRegisterConnectionIDs removes the connectionID from the map
func (t *tcpServer) unregisterConnectionIDs(pkt *packets.DisconnectPacket, conn net.Conn) {
	t.mu.Lock()
	delete(t.clients, pkt.ConnectionID)
	t.mu.Unlock()
}
