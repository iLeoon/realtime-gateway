package packets

import (
	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ConstructPacket creates and returns the appropriate packet instance for
// the given opcode. It acts as a factory that maps each opcode to its
// corresponding concrete packet type. Internally, the function allocates
// the specific packet struct and returns it as a BuildPayload interface.
func ConstructPacket(ope uint8) (BuildPayload, error) {
	const path errors.PathName = "protocol/packets/factor"
	const op errors.Op = "packets.ConstructPacket"
	switch ope {
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
	case UPDATE_MESSAGE:
		return &UpdateMessagePacket{}, nil
	case UPDATE_RESPONSE:
		return &ResponseUpdateMessagePacket{}, nil
	case DELETE_MESSAGE:
		return &DeleteMessagePacket{}, nil
	case DELETE_RESPONSE:
		return &ResponseDeleteMessagePacket{}, nil
	case ERROR:
		return &ErrorPacket{}, nil
	case TYPING:
		return &TypingPacket{}, nil
	case TYPING_RESPONSE:
		return &ResponseTypingPacket{}, nil
	case PRESENCE_RESPONSE:
		return &ResponsePresencePacket{}, nil
	}

	return nil, errors.B(path, op, errors.Internal, "unknown packet type")

}
