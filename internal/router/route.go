package router

import (
	"github.com/iLeoon/chatserver/internal/websocket"
	"github.com/iLeoon/chatserver/pkg/protcol"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

type Router struct {
	clients *map[uint32]*websocket.Client
}

func NewRouter(clients *map[uint32]*websocket.Client) *Router {
	return &Router{
		clients: clients,
	}
}

func (r *Router) Route(f *protcol.Frame) {

	switch pkt := f.Payload.(type) {
	case *packets.ResponseMessagePacket:
		r.handleResponseMessage(pkt)

	}
}

func (r *Router) handleResponseMessage(pkt *packets.ResponseMessagePacket) {

	recipient := pkt.ToConnectionID

	client := (*r.clients)[recipient]

	client.Send <- []byte(pkt.ResContent)

}
