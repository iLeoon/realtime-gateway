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
	return DELETE_RESPONSE
}

func (d *ResponseDeleteMessagePacket) Encode() ([]byte, error) {
	const path errors.PathName = "packets/delete_response"
	const op errors.Op = "ResponseDeleteMessagePacket.Encode"
	b := make([]byte, 12)

	binary.BigEndian.PutUint32(b[:4], d.MessageID)
	binary.BigEndian.PutUint32(b[4:8], d.ConversationID)
	binary.BigEndian.PutUint32(b[8:12], d.AuthorID)

	return b, nil
}

func (u *ResponseDeleteMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/delet_response"
	const op errors.Op = "ResponseDeleteMessagePacket.Decode"

	if len(b) < 12 {
		return errors.B(path, op, errors.Client, "delete response message packet length can't be less than 8")
	}

	u.MessageID = binary.BigEndian.Uint32(b[:4])
	if u.MessageID == 0 {
		return errors.B(path, op, "MessageID field is empty or 0")
	}

	u.ConversationID = binary.BigEndian.Uint32(b[4:8])
	if u.MessageID == 0 {
		return errors.B(path, op, "ConversationID field is empty or 0")
	}

	u.AuthorID = binary.BigEndian.Uint32(b[8:12])
	if u.MessageID == 0 {
		return errors.B(path, op, "Author field is empty or 0")
	}
	return nil
}
