package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/pkg/protcol/errors"
)

// DisconnectPacket indicates that the client intends to terminate the
// session gracefully.
type DisconnectPacket struct {
	ConnectionID uint32
}

func (d *DisconnectPacket) String() string {
	return fmt.Sprintf("DisconnectPacket{ConnectionID: %d}", d.ConnectionID)
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

	if d.ConnectionID == 0 {
		return fmt.Errorf("The connectionID cannot be 0: %w", errors.ErrPktSize)

	}

	return nil

}
