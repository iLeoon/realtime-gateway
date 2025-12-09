package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/pkg/protocol/errors"
)

// ConnectPacket represents a connection request sent by a client when it
// initially joins the system.
type ConnectPacket struct {
	ConnectionID uint32 // ConnectionID is a unique identifier for the connecting client.
}

func (c *ConnectPacket) String() string {
	return fmt.Sprintf("ConnectPacket{ConnectionID: %d}", c.ConnectionID)
}

// Type returns the opcode.
func (c *ConnectPacket) Type() uint8 {
	return CONNECT
}

// Encode serializes the packet fields into a payload.
func (c *ConnectPacket) Encode() ([]byte, error) {
	payloadSlice := make([]byte, 4)

	binary.BigEndian.PutUint32(payloadSlice, c.ConnectionID)

	return payloadSlice, nil

}

// Decode parses the payload and fills the struct.
func (c *ConnectPacket) Decode(b []byte) error {
	if len(b) != 4 {
		return fmt.Errorf("%w", errors.ErrPktSize)
	}
	c.ConnectionID = binary.BigEndian.Uint32(b)
	if c.ConnectionID == 0 {
		return fmt.Errorf("The connectionID cannot be 0: %w", errors.ErrPktSize)

	}
	return nil
}
