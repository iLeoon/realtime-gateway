package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

type ResponseDeleteMessagePacket struct {
	MessageID      uint32
	ConversationID uint32
	AuthorID       uint32
}

func (d *ResponseDeleteMessagePacket) String() string {
	return fmt.Sprintf("ResponseDeleteMessagePacket{MessageID: %d, ConversationID: %d, AuthorID: %d}", d.MessageID, d.ConversationID, d.AuthorID)
}

func (d *ResponseDeleteMessagePacket) Type() uint8 {
	return DeleteResponse
}

func (d *ResponseDeleteMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 12)

	binary.BigEndian.PutUint32(b[:4], d.MessageID)
	binary.BigEndian.PutUint32(b[4:8], d.ConversationID)
	binary.BigEndian.PutUint32(b[8:12], d.AuthorID)

	return b, nil
}

func (d *ResponseDeleteMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/delet_response"
	const op errors.Op = "ResponseDeleteMessagePacket.Decode"

	if len(b) < 12 {
		return errors.B(path, op, errors.Client, "delete response message packet length can't be less than 8")
	}

	d.MessageID = binary.BigEndian.Uint32(b[:4])
	if d.MessageID == 0 {
		return errors.B(path, op, "MessageID field is empty or 0")
	}

	d.ConversationID = binary.BigEndian.Uint32(b[4:8])
	if d.MessageID == 0 {
		return errors.B(path, op, "ConversationID field is empty or 0")
	}

	d.AuthorID = binary.BigEndian.Uint32(b[8:12])
	if d.MessageID == 0 {
		return errors.B(path, op, "Author field is empty or 0")
	}
	return nil
}
