package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/chatserver/pkg/protcol/errors"
)

type DisconnectPacket struct {
	ConnectionID uint32
}

func (d *DisconnectPacket) Type() uint8 {
	return DISCONNECT
}

func (d *DisconnectPacket) Encode() ([]byte, error) {

	payloadSlice := make([]byte, 4)

	binary.BigEndian.PutUint32(payloadSlice, d.ConnectionID)
	return payloadSlice, nil

}

func (d *DisconnectPacket) Decode(b []byte) error {

	if len(b) != 4 {
		return fmt.Errorf("%w", errors.ErrPktSize)
	}
	d.ConnectionID = binary.BigEndian.Uint32(b)

	return nil

}
