package router

import (
	"encoding/json"
	"fmt"

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
func (r *router) Route(pkt packets.BuildPayload, userID string, connectionID uint32) {

	switch p := pkt.(type) {
	case *packets.ResponseMessagePacket:
		r.handleResponseMessage(p, userID, connectionID)
	case *packets.ResponseUpdateMessagePacket:
		r.handleResponseUpdateMessage(p, userID, connectionID)
	case *packets.ResponseDeleteMessagePacket:
		r.handleResponseDeleteMessage(p, userID, connectionID)
	case *packets.ResponseTypingPacket:
		r.handleResponseTyping(p, userID, connectionID)
	case *packets.ResponsePresencePacket:
		r.handleResponsePresence(p, userID, connectionID)
	default:
	}
}

// handleResponseMessage delivers a ResponseMessagePacket to its intended
// WebSocket recipient.
func (r *router) handleResponseMessage(pkt *packets.ResponseMessagePacket, userID string, connectionID uint32) {
	var res = ResponseMessage{
		AuthorID:       pkt.AuthorID,
		ConversationID: pkt.ConversationID,
		MessageID:      pkt.MessageID,
		Content:        pkt.ResContent,
	}
	payload, err := json.Marshal(res)
	if err != nil {
		log.Error.Printf("failed to encode packet frame: %v to json", pkt)
		return
	}

	if err := r.router.Send(userID, connectionID, payload); err != nil {
		log.Error.Println("couldn't find the client", "Error", err)
		return
	}
}

// handleupdateresponsemessage delivers an updated message to its intended
// websocket recipient.
func (r *router) handleResponseUpdateMessage(pkt *packets.ResponseUpdateMessagePacket, userID string, connectionID uint32) {
	var res = ResponseUpdateMessage{
		ConversationID: pkt.ConversationID,
		MessageID:      pkt.MessageID,
		UpdatedAt:      pkt.Updated_at,
		Content:        pkt.ResContent,
	}
	payload, err := json.Marshal(res)
	if err != nil {
		log.Error.Printf("failed to encode packet frame: %v to json", pkt)
		return
	}

	if err := r.router.Send(userID, connectionID, payload); err != nil {
		log.Error.Println("couldn't find the client", "error", err)
		return
	}
}

// handleResponseTyping delivers a typing indicator to its intended WebSocket recipient.
func (r *router) handleResponseTyping(pkt *packets.ResponseTypingPacket, userID string, connectionID uint32) {
	res := ResponseTyping{
		ConversationID: pkt.ConversationID,
		UserID:         fmt.Sprintf("%d", pkt.UserID),
		IsTyping:       pkt.IsTyping,
	}
	payload, err := json.Marshal(res)
	if err != nil {
		log.Error.Printf("failed to encode typing packet: %v to json", pkt)
		return
	}

	if err := r.router.Send(userID, connectionID, payload); err != nil {
		log.Error.Println("couldn't find the client", "error", err)
		return
	}
}

// handleResponseDeleteMessage delivers a delete notification to its intended WebSocket recipient.
func (r *router) handleResponseDeleteMessage(pkt *packets.ResponseDeleteMessagePacket, userID string, connectionID uint32) {
	var res = ResponseDeleteMessage{
		MessageID:      pkt.MessageID,
		ConversationID: pkt.ConversationID,
		AuthorID:       pkt.AuthorID,
	}
	payload, err := json.Marshal(res)
	if err != nil {
		log.Error.Printf("failed to encode packet frame: %v to json", pkt)
		return
	}

	if err := r.router.Send(userID, connectionID, payload); err != nil {
		log.Error.Println("couldn't find the client", "error", err)
		return
	}
}

// handleResponsePresence delivers a presence notification to its intended WebSocket recipient.
func (r *router) handleResponsePresence(pkt *packets.ResponsePresencePacket, userID string, connectionID uint32) {
	res := ResponsePresence{
		UserID:   fmt.Sprintf("%d", pkt.UserID),
		IsOnline: pkt.IsOnline,
	}
	payload, err := json.Marshal(res)
	if err != nil {
		log.Error.Printf("failed to encode presence packet: %v to json", pkt)
		return
	}

	if err := r.router.Send(userID, connectionID, payload); err != nil {
		log.Error.Println("couldn't find the client", "error", err)
		return
	}
}
