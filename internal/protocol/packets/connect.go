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
	UserID       uint32 // UserID is the authenticated user who owns this connection.
}

func (c *ConnectPacket) String() string {
	return fmt.Sprintf("ConnectPacket{ConnectionID: %d, UserID: %d}", c.ConnectionID, c.UserID)
}

// Type returns the opcode.
func (c *ConnectPacket) Type() uint8 {
	return Connect
}

// Encode serializes the packet fields into a payload.
// Layout: [4 bytes ConnectionID][4 bytes UserID]
func (c *ConnectPacket) Encode() ([]byte, error) {
	b := make([]byte, 8)

	binary.BigEndian.PutUint32(b[:4], c.ConnectionID)
	binary.BigEndian.PutUint32(b[4:8], c.UserID)

	return b, nil
}

// Decode parses the payload and fills the struct.
func (c *ConnectPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/connect"
	const op errors.Op = "ConnectPacket.Decode"
	if len(b) != 8 {
		return errors.B(path, op, errors.Internal, "invalid packet fields size")
	}
	c.ConnectionID = binary.BigEndian.Uint32(b[:4])
	if c.ConnectionID == 0 {
		return errors.B(path, op, errors.Internal, "connectionID field is empty or 0")
	}
	c.UserID = binary.BigEndian.Uint32(b[4:8])
	if c.UserID == 0 {
		return errors.B(path, op, errors.Internal, "userID field is empty or 0")
	}

	return nil
}
