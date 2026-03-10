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
	case SendMessage:
		return &SendMessagePacket{}, nil
	case ResponseMessage:
		return &ResponseMessagePacket{}, nil
	case Connect:
		return &ConnectPacket{}, nil
	case Disconnect:
		return &DisconnectPacket{}, nil
	case Ping:
		return &PingPacket{}, nil
	case Pong:
		return &PongPacket{}, nil
	case UpdateMessage:
		return &UpdateMessagePacket{}, nil
	case UpdateResponse:
		return &ResponseUpdateMessagePacket{}, nil
	case DeleteMessage:
		return &DeleteMessagePacket{}, nil
	case DeleteResponse:
		return &ResponseDeleteMessagePacket{}, nil
	case Error:
		return &ErrorPacket{}, nil
	case Typing:
		return &TypingPacket{}, nil
	case TypingResponse:
		return &ResponseTypingPacket{}, nil
	case PresenceResponse:
		return &ResponsePresencePacket{}, nil
	}

	return nil, errors.B(path, op, errors.Internal, "unknown packet type")

}
