package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ResponseMessagePacket represents a message delivered from the server
// back to a client.
type ResponseMessagePacket struct {
	ToConnectionID uint32
	ResContent     string
}

func (r *ResponseMessagePacket) String() string {
	return fmt.Sprintf("ResponseMessagePacket{ToConnectionID: %d, ResContent: %q}", r.ToConnectionID, r.ResContent)
}

func (r *ResponseMessagePacket) Type() uint8 {
	return RESPONSE_MESSAGE
}

func (r *ResponseMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 4+len(r.ResContent))
	binary.BigEndian.PutUint32(b[:4], r.ToConnectionID)
	copy(b[4:], r.ResContent)
	return b, nil
}

func (r *ResponseMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/response_message"
	const op errors.Op = "ResponseMessagePacket.Decode"

	if len(b) < 4 {
		return errors.B(path, op, errors.Client, "response message packet length can't be less than 4")
	}

	r.ToConnectionID = binary.BigEndian.Uint32(b[:4])
	if r.ToConnectionID == 0 {
		return errors.B(path, op, errors.Client, "connectionID field is empty or 0")
	}

	if len(b[4:]) > 512 {
		return errors.B(path, op, errors.Client, fmt.Errorf("message size(%v) hit the maximum size", len(b[4:])))
	}
	if len(b[4:]) == 0 {
		return errors.B(path, op, errors.Client, "message field is empty")
	}

	r.ResContent = string(b[4:])
	return nil
}
