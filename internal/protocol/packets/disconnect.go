package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// DisconnectPacket indicates that the client intends to terminate the
// session gracefully.
type DisconnectPacket struct {
	ConnectionID uint32
	UserID       uint32
}

func (d *DisconnectPacket) String() string {
	return fmt.Sprintf("DisconnectPacket{ConnectionID: %d, UserID: %d}", d.ConnectionID, d.UserID)
}

func (d *DisconnectPacket) Type() uint8 {
	return DISCONNECT
}

// Encode serializes the packet fields into a payload.
// Layout: [4 bytes ConnectionID][4 bytes UserID]
func (d *DisconnectPacket) Encode() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b[:4], d.ConnectionID)
	binary.BigEndian.PutUint32(b[4:8], d.UserID)
	return b, nil
}

func (d *DisconnectPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/disconnect"
	const op errors.Op = "DisconnectPacket.Decode"

	if len(b) != 8 {
		return errors.B(path, op, errors.Internal, "invalid packet fields size")
	}
	d.ConnectionID = binary.BigEndian.Uint32(b[:4])
	if d.ConnectionID == 0 {
		return errors.B(path, op, errors.Internal, "connectionID field is empty or 0")
	}
	d.UserID = binary.BigEndian.Uint32(b[4:8])
	if d.UserID == 0 {
		return errors.B(path, op, errors.Internal, "userID field is empty or 0")
	}

	return nil
}
