package packets

import (
	"bytes"
	"encoding/binary"

	"github.com/iLeoon/chatserver/pkg/logger"
)

type ConnectPacket struct {
	ConnectionID uint32
}

func (c *ConnectPacket) Type() uint8 {
	return CONNECT
}

func (c *ConnectPacket) Encode() (error, []byte) {
	var body bytes.Buffer

	if err := binary.Write(&body, binary.BigEndian, c.ConnectionID); err != nil {
		logger.Error("Error encoding the connectionID into the buffer", "Error", err)
		return err, nil
	}

	return nil, body.Bytes()

}

func (c *ConnectPacket) Decode(b []byte) error {
	connetionID := binary.BigEndian.Uint32(b[:4])

	c.ConnectionID = connetionID

	return nil

}
