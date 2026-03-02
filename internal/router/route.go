package router

import (
	"encoding/json"

	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
	"github.com/iLeoon/realtime-gateway/pkg/log"
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
	case *packets.ErrorPacket:
		r.handleErrorMessage(p, userID)
	}
}

// handleResponseMessage delivers a ResponseMessagePacket to its intended
// WebSocket recipient.
func (r *router) handleResponseMessage(pkt *packets.ResponseMessagePacket, userID string) {
	var res = ResponseMessage{
		AuthorID:       pkt.AuthorID,
		ConversationID: pkt.ConversationID,
		Content:        pkt.ResContent,
	}
	payload, err := json.Marshal(res)
	if err != nil {
		log.Error.Printf("failed to encode packet frame: %v to json", pkt)
		return
	}
	connectionID := pkt.ToConnectionID

	if err := r.router.Send(userID, connectionID, payload); err != nil {
		log.Error.Println("couldn't find the client", "Error", err)
		return
	}

}

func (r *router) handleErrorMessage(pkt *packets.ErrorPacket, userID string) {
	message := []byte(pkt.Message)
	if err := r.router.Send(userID, pkt.ConnectionID, message); err != nil {
		log.Error.Println("failed to deliver error to client", err)
	}
}
