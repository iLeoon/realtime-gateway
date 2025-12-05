package packets

import (
	"github.com/iLeoon/chatserver/pkg/logger"
)

func ConstructPacket(op uint8) BuildPayload {
	switch op {
	case SEND_MESSAGE:
		return &SendMessage{}
	case RESPONSE_MESSAGE:
		return &ResponseMessagePacket{}
	case CONNECT:
		return &ConnectPacket{}
	case DISCONNECT:
		return &DisconnectPacket{}
	default:
		logger.Error("This packet type does not exist")
		return nil
	}
}
