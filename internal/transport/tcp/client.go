package tcp

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/protocol"
	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

type Router interface {
	Route(p packets.BuildPayload, userId string, connectionID uint32)
}

type Signaler interface {
	Signal(userId string, connectionId uint32, code int, reason string)
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
	conn, err := net.Dial("tcp", t.config.TcpPort)
	if err != nil {
		return nil, err
	}

	log.Info.Println("The tcp client successfully established a connection between websocket gateway and tcp server")

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
	const path errors.PathName = "tcp/client"
	const op errors.Op = "tcpClient.WriteToServer"
	var pkt packets.BuildPayload
	cp := &ClientPayload{}

	// Unmarshal the incmoing byets from the gateway to the client payload struct
	err := json.Unmarshal(data, cp)
	if err != nil {
		return errors.B(path, op, errors.Client, err)
	}

	// Check the message type (opcode) based on the ClientPayload.Opcode
	// then build the packet and pass to writePacket to send it
	// into the raw tcp connection
	switch cp.Opcode {
	case "send_message":
		// Build the readable packet.
		var data SendMessagePayload
		err := json.Unmarshal(cp.Payload, &data)
		if err != nil {
			return errors.B(path, op, err, errors.Client)
		}
		convIDToInt, err := strconv.ParseUint(data.ConversationID, 10, 32)
		if err != nil {
			return errors.B(path, op, errors.Client, err)
		}

		pkt = &packets.SendMessagePacket{
			ConversationID: uint32(convIDToInt),
			Content:        data.Content,
		}
	case "update_message":
		var data UpdateMessagePayload
		err := json.Unmarshal(cp.Payload, &data)
		if err != nil {
			return errors.B(path, op, err, errors.Client)
		}

		convIDToInt, err := strconv.ParseUint(data.ConversationID, 10, 32)
		if err != nil {
			return errors.B(path, op, errors.Client, err)
		}

		messageIDToInt, err := strconv.ParseUint(data.MessageID, 10, 32)
		if err != nil {
			return errors.B(path, op, errors.Client, err)
		}

		pkt = &packets.UpdateMessagePacket{
			ConversationID: uint32(convIDToInt),
			MessageID:      uint32(messageIDToInt),
			Content:        data.Content,
		}

	case "delete_message":
		var data DeleteMessagePayload
		err := json.Unmarshal(cp.Payload, &data)
		if err != nil {
			return errors.B(path, op, err, errors.Client)
		}

		messageIDToInt, err := strconv.ParseUint(data.MessageID, 10, 32)
		if err != nil {
			return errors.B(path, op, errors.Client, err)
		}

		convIDToInt, err := strconv.ParseUint(data.ConversationID, 10, 32)
		if err != nil {
			return errors.B(path, op, errors.Client, err)
		}
		pkt = &packets.DeleteMessagePacket{
			MessageID:      uint32(messageIDToInt),
			ConversationID: uint32(convIDToInt),
		}

	case "typing":
		var data TypingPayload
		err := json.Unmarshal(cp.Payload, &data)
		if err != nil {
			return errors.B(path, op, err, errors.Client)
		}

		convIDToInt, err := strconv.ParseUint(data.ConversationID, 10, 32)
		if err != nil {
			return errors.B(path, op, errors.Client, err)
		}

		pkt = &packets.TypingPacket{
			ConversationID: uint32(convIDToInt),
			IsTyping:       data.IsTyping,
		}
	default:
		return errors.B(path, op, errors.Client, errors.Errorf("Invalid packet type %T", cp.Opcode))
	}

	if err := t.writePacket(pkt); err != nil {
		return errors.B(path, op, err)
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
	const path errors.PathName = "tcp/client"
	const op errors.Op = "tcpClient.ReadFromServer"
	wsCode := 1011 // default: internal server error
	var reason string
	defer func() {
		t.signal.Signal(t.userID, t.connectionID, wsCode, reason)
		t.conn.Close()
		log.Info.Printf("%q: %q: tcp client terminated it's connection", path, op)
	}()

	for {
		// Decode the frame.
		if err := t.conn.SetReadDeadline(time.Now().Add(readDuration)); err != nil {
			wrappedErr := errors.B(path, op, err)
			log.Error.Println("failed to read the packet from the peer", wrappedErr)
			return
		}
		frame, err := protocol.DecodeFrame(t.conn)
		if err != nil {
			// Error that gives context to the rest of the logs
			wrappedErr := errors.B(path, op, err)
			switch {
			case errors.Is(err, io.EOF):
				log.Info.Println("TCP connection is closed by peer(server)", wrappedErr)
				wsCode, reason = 1006, "server closed connection"
				return
			case errors.Is(err, net.ErrClosed):
				wsCode, reason = 1000, "connection is closed"
				return
			case errors.Is(err, os.ErrDeadlineExceeded):
				wsCode, reason = 1006, "unexpected failure please refresh"
				log.Info.Println("read deadline exceeded, closing connection", wrappedErr)
				return
			default:
				log.Error.Println("unexpected error from the tcp server", wrappedErr)
				wsCode, reason = 1011, "connection is terminated"
				return
			}
		}

		switch pkt := frame.Payload.(type) {
		case *packets.PingPacket:
			if err := t.pongRes(); err != nil {
				errorWrapper := errors.B(path, op, err)
				log.Error.Println("pong packet failed", errorWrapper)
				wsCode, reason = 1006, "unexpected failure please refresh"
				return
			}
		case *packets.ErrorPacket:
			if pkt.Code == errors.Client {
				wsCode, reason = 1008, pkt.Message // protocol error
				return
			}

		case *packets.ResponseMessagePacket:
			t.router.Route(pkt, t.userID, t.connectionID)
		case *packets.ResponseUpdateMessagePacket:
			t.router.Route(pkt, t.userID, t.connectionID)
		case *packets.ResponseDeleteMessagePacket:
			t.router.Route(pkt, t.userID, t.connectionID)
		case *packets.ResponseTypingPacket:
			t.router.Route(pkt, t.userID, t.connectionID)
		case *packets.ResponsePresencePacket:
			t.router.Route(pkt, t.userID, t.connectionID)
		}
		log.Info.Println("Decode packet", "packet", frame.Payload.String())
	}
}

// OnConnect is invoked as part of the session lifecycle when a new
// WebSocket client successfully connects.
//
// It constructs the connectp packet encodes it into a frame, and forwards it
// to the TCP server so the engine can register and track the new
// client session.
func (t *tcpClient) OnConnect() error {
	const path errors.PathName = "tcp/client"
	const op errors.Op = "tcpClient.onConnect"

	// userID originates as a string from the JWT claims but is represented
	// as uint32 internally to avoid repeated string conversions.
	userIDToInt, err := strconv.ParseUint(t.userID, 10, 32)
	if err != nil {
		return errors.B(path, op, errors.Client, "faild to convert userID to int", err)

	}
	pkt := &packets.ConnectPacket{
		ConnectionID: t.connectionID,
		UserID:       uint32(userIDToInt),
	}
	if err := t.writePacket(pkt); err != nil {
		return errors.B(path, op, err)
	}
	return nil
}

// DisConnect is basically the inverse of OnConnect.
func (t *tcpClient) OnDisConnect() error {
	const path errors.PathName = "tcp/client"
	const op errors.Op = "tcpClient.onDisconnect"

	userIDToInt, err := strconv.ParseUint(t.userID, 10, 32)
	if err != nil {
		return errors.B(path, op, errors.Client, "faild to convert userID to int", err)
	}
	pkt := &packets.DisconnectPacket{
		ConnectionID: t.connectionID,
		UserID:       uint32(userIDToInt),
	}
	if err := t.writePacket(pkt); err != nil {
		return errors.B(path, op, err)
	}
	return nil
}

func (t *tcpClient) pongRes() error {
	const path errors.PathName = "tcp/client"
	const op errors.Op = "tcpClient.pongRes"
	pkt := &packets.PongPacket{}
	if err := t.writePacket(pkt); err != nil {
		return errors.B(path, op, err)
	}
	return nil
}

// writePacket constructs the frame and write it
// into the raw tcp connection
func (t *tcpClient) writePacket(pkt packets.BuildPayload) error {
	const path errors.PathName = "tcp/client"
	const op errors.Op = "tcpClient.writePacket"
	if err := t.conn.SetWriteDeadline(time.Now().Add(writeDuration)); err != nil {
		return errors.B(path, op, "connection is unhealthy", err)
	}

	// Construct the frame, encode it, and then send it to the TCP server.
	frame := protocol.ConstructFrame(pkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return errors.B(path, op, err)
	}
	return nil
}
