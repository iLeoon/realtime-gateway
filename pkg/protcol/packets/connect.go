package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/chatserver/pkg/protcol/errors"
)

type ConnectPacket struct {
	ConnectionID uint32
}

func (c *ConnectPacket) Type() uint8 {
	return CONNECT
}

func (c *ConnectPacket) Encode() ([]byte, error) {
	payloadSlice := make([]byte, 4)

	binary.BigEndian.PutUint32(payloadSlice, c.ConnectionID)

	return payloadSlice, nil

}

func (c *ConnectPacket) Decode(b []byte) error {
	if len(b) != 4 {
		return fmt.Errorf("%w", errors.ErrPktSize)
	}
	c.ConnectionID = binary.BigEndian.Uint32(b)

	return nil

}
