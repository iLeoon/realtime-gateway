package tcpclient

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/internal/router"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/protcol"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

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
	conn   net.Conn
	router *router.Router // Router routes the data coming from Tcp server to websocket gateway.
}

// NewTcpClient establishes the TCP connection between the WebSocket
// gateway and the TCP engine. This function is invoked from the
// WebSocket server during startup to create the bride
// and be to send/receive messages.
func NewTCPClient(conf *config.Config, router *router.Router) *tcpClient {
	conn, err := net.Dial("tcp", conf.TCPServer.Port)
	if err != nil {
		logger.Error("Can't connect to the tcp server", "Error", err)
	}

	logger.Info("The websocket gateway successfully connected to the tcp server")

	client := &tcpClient{
		conn:   conn,
		router: router,
	}
	go client.ReadFromServer()
	return client

}

// ReadFromGateway handles incoming messages from the browser/WebSocket
// client. It receives raw JSON payloads, unmarshals them into the
// ClientPayload structure, and uses the opcode to determine which internal
// packet type to construct.
//
// Based on the opcode, the method builds the appropriate packet
// encodes it into a protocol frame, and transmits it
// to the TCP engine using the underlying TCP connection.
func (t *tcpClient) ReadFromGateway(data []byte, connectionID uint32) error {
	cp := &ClientPayload{}

	// Unmarshal the incmoing byets from the gateway to the client payload struct
	json.Unmarshal(data, cp)

	// Check the message type (opcode) based on the ClientPayload.Opcode.
	switch cp.Opcode {
	case "send_message":

		// Build the readable packet.
		var data SendMessagePayload
		json.Unmarshal(cp.Payload, &data)
		pkt := &packets.SendMessagePacket{
			ConnectionID: connectionID,
			Content:      data.Content,
		}

		// Construct the frame, encode it, and then send it to the TCP server.
		frame := protcol.ConstructFrame(pkt)
		err := frame.EncodeFrame(t.conn)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid packet type %s", cp.Opcode)
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
		t.conn.Close()
		logger.Info("Gateway closed connection to tcp server")
	}()
	for {
		// Decode the frame.
		frame, err := protcol.DecodeFrame(t.conn)

		if err != nil {
			logger.Error("Invalid incoming data from tcp server", "Error", err)
			return
		}

		// Push it to the router.
		t.router.Route(frame)

		logger.Debug("Decode packet", "packet", frame.Payload.String())

	}
}

// OnConnect is invoked as part of the session lifecycle when a new
// WebSocket client successfully connects.
//
// It constructs the connectp packet encodes it into a frame, and forwards it
// to the TCP server so the engine can register and track the new
// client session.
func (t *tcpClient) OnConnect(connectionID uint32) error {
	pkt := &packets.ConnectPacket{
		ConnectionID: connectionID,
	}

	frame := protcol.ConstructFrame(pkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return err
	}
	return nil
}

// DisConnect is basically the inverse of OnConnect.
func (t *tcpClient) DisConnect(connectionID uint32) error {
	pkt := &packets.DisconnectPacket{
		ConnectionID: connectionID,
	}

	frame := protcol.ConstructFrame(pkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return err
	}
	return nil
}
