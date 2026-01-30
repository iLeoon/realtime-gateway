package tcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/protocol"
	"github.com/iLeoon/realtime-gateway/pkg/protocol/packets"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

type Router interface {
	Route(p packets.BuildPayload, userId string)
}

type Signaler interface {
	Signal(userId string, connectionId uint32)
}

// TcpClient acts as the transporter between the WebSocket gateway and the
// TCP engine. It implements the Session interface.
//
// It wraps the underlying TCP connection and is responsible
// for transmitting encoded frames to the server as well as receiving frame
// data from the TCP engine and routing it back to the WebSocket gateway.
//
// The TcpClient therefore forms the low-level communication layer that
// bridges user-facing WebSocket sessions with backend protocol logic.
//
// The flow is:
//
//	Browser JSON → ReadFromGateway → ConstructPacket → EncodeFrame → TCP Engine
//	TCP Engine  → DecodeFrame → Route Packet → Handle Response → WebSocket Client
type tcpClient struct {
	conn         net.Conn
	config       *config.Config
	router       Router
	connectionID uint32 // the requester connection ID
	userID       string // the requester user ID
	signal       Signaler
}

type tcpClientFactory struct {
	config *config.Config
	router Router // Router routes the data coming from Tcp server to websocket gateway.
	signal Signaler
}

func NewFactory(c *config.Config, r Router, s Signaler) *tcpClientFactory {
	return &tcpClientFactory{
		config: c,
		router: r,
		signal: s,
	}
}

// NewTCPClient establishes the TCP connection between the WebSocket
// gateway and the TCP engine. This function create the bridge
// between the websocket gateway and tcp server
// to send/receive messages.
func (t *tcpClientFactory) NewClient(userID string, connectionID uint32) (session.Session, error) {
	conn, err := net.Dial("tcp", t.config.TCP.TcpPort)
	if err != nil {
		return nil, err

	}

	logger.Info("The tcp client successfully established a connection between websocket gateway and tcp server")

	client := &tcpClient{
		conn:         conn,
		config:       t.config,
		router:       t.router,
		signal:       t.signal,
		userID:       userID,
		connectionID: connectionID,
	}
	go client.ReadFromServer()
	return client, nil
}

// ReadFromGateway handles incoming messages from the browser/WebSocket gateway
// client. It receives raw JSON payloads, unmarshals them into the
// ClientPayload structure, and uses the opcode to determine which internal
// packet type to construct.
//
// Based on the opcode, the method builds the appropriate packet
// encodes it into a protocol frame, and transmits it
// to the TCP engine using the underlying TCP connection.
func (t *tcpClient) WriteToServer(data []byte) error {
	var pkt packets.BuildPayload
	cp := &ClientPayload{}

	// Unmarshal the incmoing byets from the gateway to the client payload struct
	err := json.Unmarshal(data, cp)
	if err != nil {
		return err
	}

	// Check the message type (opcode) based on the ClientPayload.Opcode.
	switch cp.Opcode {
	case "send_message":
		// Build the readable packet.
		var data SendMessagePayload
		err := json.Unmarshal(cp.Payload, &data)
		if err != nil {
			return err
		}
		pkt = &packets.SendMessagePacket{
			ConnectionID: t.connectionID,
			Content:      data.Content,
		}
	default:
		return fmt.Errorf("Invalid packet type %s", cp.Opcode)
	}

	if err := t.connectionHygiene(pkt); err != nil {
		t.conn.Close()
		return err
	}
	return nil
}

// ReadFromServer performs the reverse operation of ReadFromGateway. It
// reads raw bytes arriving from the TCP engine, decodes them into their
// human-readable packet/struct form, and forwards the resulting frame to
// the router for delivery back to the WebSocket gateway.
//
// This method must run inside its own goroutine because it performs
// blocking I/O while waiting for incoming data from the TCP connection.
func (t *tcpClient) ReadFromServer() {
	defer func() {
		t.signal.Signal(t.userID, t.connectionID)
		t.conn.Close()
		logger.Info("Gateway closed connection to tcp server")
	}()

	for {
		// Decode the frame.
		t.conn.SetReadDeadline(time.Now().Add(readDuration))
		frame, err := protocol.DecodeFrame(t.conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("TCP connection is closed by peer(server)")
				return
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}

			logger.Error("Unexpected tcp read error", "error", err)
			return
		}

		switch pkt := frame.Payload.(type) {
		case *packets.PingPacket:
			t.pongRes()
		default:
			// Push it to the router.
			t.router.Route(pkt, t.userID)
		}

		logger.Debug("Decode packet", "packet", frame.Payload.String())
	}
}

// OnConnect is invoked as part of the session lifecycle when a new
// WebSocket client successfully connects.
//
// It constructs the connectp packet encodes it into a frame, and forwards it
// to the TCP server so the engine can register and track the new
// client session.
func (t *tcpClient) OnConnect() error {
	pkt := &packets.ConnectPacket{
		ConnectionID: t.connectionID,
	}
	if err := t.connectionHygiene(pkt); err != nil {
		return err
	}
	return nil
}

// DisConnect is basically the inverse of OnConnect.
func (t *tcpClient) OnDisConnect() error {
	pkt := &packets.DisconnectPacket{
		ConnectionID: t.connectionID,
	}
	if err := t.connectionHygiene(pkt); err != nil {
		return err
	}

	return nil
}

func (t *tcpClient) connectionHygiene(pkt packets.BuildPayload) error {
	if err := t.conn.SetWriteDeadline(time.Now().Add(writeDuration)); err != nil {
		return fmt.Errorf("connection is unhealty: %w", err)
	}

	// Construct the frame, encode it, and then send it to the TCP server.
	frame := protocol.ConstructFrame(pkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return err
	}
	return nil
}

func (t *tcpClient) pongRes() {
	pkt := &packets.PongPacket{}
	t.connectionHygiene(pkt)
}
