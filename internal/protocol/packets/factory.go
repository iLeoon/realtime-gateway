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
	}

	return nil, errors.B(path, op, errors.Internal, "unknown packet type")

}
