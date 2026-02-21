package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
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
	const path errors.PathName = "packets/connect"
	const op errors.Op = "ConnectPacket.Decode"
	if len(b) != 4 {
		return errors.B(path, op, errors.Internal, "invalid packet field size")
	}
	c.ConnectionID = binary.BigEndian.Uint32(b)
	if c.ConnectionID == 0 {
		return errors.B(path, op, errors.Internal, "invalid packet field size")
	}
	return nil
}
