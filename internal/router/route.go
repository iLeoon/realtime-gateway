package router

import (
	"github.com/iLeoon/realtime-gateway/internal/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/protocol"
	"github.com/iLeoon/realtime-gateway/pkg/protocol/packets"
)

// Router forwards decoded packets from the TCP engine to the appropriate
// WebSocket clients. It holds a reference to the active client map
// maintained by the WebSocket server and uses it to deliver outbound
// messages to their intended recipients.
type Router struct {
	clients *map[uint32]*websocket.Client
}

// Create a new router.
func NewRouter(clients *map[uint32]*websocket.Client) *Router {
	return &Router{
		clients: clients,
	}
}

// Route receives a decoded protocol frame from the TCP engine and
// dispatches it to the appropriate handler based on the concrete packet
// type stored inside the frame’s payload.
func (r *Router) Route(f *protocol.Frame) {
	switch pkt := f.Payload.(type) {
	case *packets.ResponseMessagePacket:
		r.handleResponseMessage(pkt)
	}
}

// handleResponseMessage delivers a ResponseMessagePacket to its intended
// WebSocket recipient.
func (r *Router) handleResponseMessage(pkt *packets.ResponseMessagePacket) {
	recipient := pkt.ToConnectionID

	client := (*r.clients)[recipient]

	client.Send <- []byte(pkt.ResContent)

}
