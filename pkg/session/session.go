package session

// Session defines the high-level connection behavior between the WebSocket
// gateway and the TCP engine. It acts as an abstraction over the underlying
// transporter (a TCP client) and exposes the lifecycle and message-flow
// operations required to bridge the two systems.
//
// A Session is responsible for:
//   - Managing connection lifecycle events (OnConnect, Disconnect)
//   - Reading messages coming from the WebSocket gateway
//   - Reading messages coming from the TCP engine
//   - Forwarding, translating, or processing packets between both sides
//
// Each concrete session implementation wraps a TCP client (the transporter)
// and coordinates bi-directional data flow between the gateway and the
// server engine.
// Read messages coming from the WebSocket gateway.
// Read messages coming from the TCP engine.
// Forward, and process packets between both sides.
type Session interface {
	// OnConnect is invoked when the WebSocket session is established.
	// Its primary responsibility is to create and send a Connect packet to the TCP engine.
	OnConnect(uint32) error

	// Disconnect is called when the WebSocket closes or the server decides
	// to terminate the session. It constructs and sends a Disconnect packet.
	DisConnect(uint32) error

	// ReadFromGateway processes incoming data from the WebSocket gateway.
	// This method typically receives encoded frames or raw messages from
	// the client, decodes them into packets, and forwards them to the TCP
	// engine through the transporter.
	ReadFromGateway(data []byte, connectionID uint32, userID string) error

	// ReadFromServer handles data arriving from the TCP engine. It reads
	// raw bytes from the transporter, decodes them into frames/packets,
	// and pushes them back to the WebSocket client through the gateway.
	ReadFromServer()
}

type InitiateSession interface {
	NewTCPClient(string, uint32) (Session, error)
}
