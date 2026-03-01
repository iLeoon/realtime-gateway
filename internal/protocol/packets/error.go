package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ErrorPacket carries a structured error back to the sender.
// Layout: [4 bytes ConnectionID][1 byte Kind][N bytes Message]
type ErrorPacket struct {
	ConnectionID uint32
	Code         errors.Kind
	Message      string
}

func (e *ErrorPacket) String() string {
	return fmt.Sprintf("ErrorPacket{ConnectionID: %d, Code: %s, Message: %q}", e.ConnectionID, e.Code, e.Message)
}

func (e *ErrorPacket) Type() uint8 {
	return ERROR
}

func (e *ErrorPacket) Encode() ([]byte, error) {
	b := make([]byte, 5+len(e.Message))
	binary.BigEndian.PutUint32(b[:4], e.ConnectionID)
	b[4] = uint8(e.Code)
	copy(b[5:], e.Message)
	return b, nil
}

func (e *ErrorPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/error"
	const op errors.Op = "ErrorPacket.Decode"
	if len(b) < 5 {
		return errors.B(path, op, errors.Client, "error packet too short")
	}
	e.ConnectionID = binary.BigEndian.Uint32(b[:4])
	e.Code = errors.Kind(b[4])
	e.Message = string(b[5:])
	return nil
}

