package packets

import (
	"fmt"

	"github.com/iLeoon/chatserver/pkg/protcol/errors"
)

//Build the payload pakcet based on the type event
//Coming from the browser and return a pointer
//To that packet otherwise it's a wrong packet type

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

	}

	return nil, fmt.Errorf("%w: the value of the packet %v", errors.ErrPacketType, op)
}
