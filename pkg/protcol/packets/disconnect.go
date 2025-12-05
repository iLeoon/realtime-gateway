package packets

import (
	"bytes"
	"encoding/binary"

	"github.com/iLeoon/chatserver/pkg/logger"
)

type DisconnectPacket struct {
	ConnectionID uint32
}

func (d *DisconnectPacket) Type() uint8 {
	return DISCONNECT
}

func (d *DisconnectPacket) Encode() (error, []byte) {
	var body bytes.Buffer

	if err := binary.Write(&body, binary.BigEndian, d.ConnectionID); err != nil {
		logger.Error("Error encoding the connectionID into the buffer", "Error", err)
		return err, nil
	}

	return nil, body.Bytes()

}

func (d *DisconnectPacket) Decode(b []byte) error {
	connetionID := binary.BigEndian.Uint32(b[:4])

	d.ConnectionID = connetionID

	return nil

}
