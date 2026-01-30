package router

import (
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/protocol/packets"
)

// Router forwards decoded packets from the TCP engine to the appropriate
// WebSocket clients. It holds a reference to the active client map
// maintained by the WebSocket server and uses it to deliver outbound
// messages to their intended recipients.

type Sender interface {
	Send(userId string, connectionID uint32, message []byte) (err error)
}

type router struct {
	router Sender
}

// Create a new router.
func New(r Sender) *router {
	return &router{
		router: r,
	}
}

// Route receives a decoded protocol frame from the TCP engine and
// dispatches it to the appropriate handler based on the concrete packet
// type stored inside the frame’s payload.
func (r *router) Route(pkt packets.BuildPayload, userID string) {
	switch p := pkt.(type) {
	case *packets.ResponseMessagePacket:
		r.handleResponseMessage(p, userID)
	}
}

// handleResponseMessage delivers a ResponseMessagePacket to its intended
// WebSocket recipient.
func (r *router) handleResponseMessage(pkt *packets.ResponseMessagePacket, userID string) {
	recipient := pkt.ToConnectionID
	message := []byte(pkt.ResContent)

	err := r.router.Send(userID, recipient, message)
	if err != nil {
		logger.Error("couldn't find the client", "Error", err)
		return
	}

}
