package packets

import (
	"fmt"

	"github.com/iLeoon/realtime-gateway/pkg/protocol/errors"
)

// ConstructPacket creates and returns the appropriate packet instance for
// the given opcode. It acts as a factory that maps each opcode to its
// corresponding concrete packet type. Internally, the function allocates
// the specific packet struct and returns it as a BuildPayload interface.
func ConstructPacket(op uint8) (BuildPayload, error) {
	switch op {
	case SEND_MESSAGE:
		return &SendMessagePacket{}, nil
	case RESPONSE_MESSAGE:
		return &ResponseMessagePacket{}, nil
	case CONNECT:
		return &ConnectPacket{}, nil
	case DISCONNECT:
		return &DisconnectPacket{}, nil
	case PING:
		return &PingPacket{}, nil
	case PONG:
		return &PongPacket{}, nil
	}

	return nil, fmt.Errorf("%w: the value of the packet %v", errors.ErrPacketType, op)
}
