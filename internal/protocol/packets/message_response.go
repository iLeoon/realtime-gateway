package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ResponseMessagePacket represents a message delivered from the server
// back to a client.
type ResponseMessagePacket struct {
	AuthorID       uint32
	ConversationID uint32
	MessageID      uint32
	ResContent     string
}

func (r *ResponseMessagePacket) String() string {
	return fmt.Sprintf("ResponseMessagePacket{AuthorID: %d, ConversationID: %d, MessageID: %d, ResContent: %q}", r.AuthorID, r.ConversationID, r.MessageID, r.ResContent)
}

func (r *ResponseMessagePacket) Type() uint8 {
	return ResponseMessage
}

func (r *ResponseMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 12+len(r.ResContent))
	binary.BigEndian.PutUint32(b[:4], r.AuthorID)
	binary.BigEndian.PutUint32(b[4:8], r.ConversationID)
	binary.BigEndian.PutUint32(b[8:12], r.MessageID)
	copy(b[12:], r.ResContent)
	return b, nil
}

func (r *ResponseMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/response_message"
	const op errors.Op = "ResponseMessagePacket.Decode"

	if len(b) < 12 {
		return errors.B(path, op, errors.Client, "response message packet length can't be less than 12")
	}

	r.AuthorID = binary.BigEndian.Uint32(b[:4])
	if r.AuthorID == 0 {
		return errors.B(path, op, errors.Client, "Author field is empty or 0")
	}

	r.ConversationID = binary.BigEndian.Uint32(b[4:8])
	if r.ConversationID == 0 {
		return errors.B(path, op, errors.Client, "conversation field is empty or 0")
	}

	r.MessageID = binary.BigEndian.Uint32(b[8:12])
	if r.MessageID == 0 {
		return errors.B(path, op, errors.Client, "messageID field is empty or 0")
	}

	if len(b[12:]) > 512 {
		return errors.B(path, op, errors.Client, fmt.Errorf("message size(%v) hit the maximum size", len(b[12:])))
	}
	if len(b[12:]) == 0 {
		return errors.B(path, op, errors.Client, "message field is empty")
	}

	r.ResContent = string(b[12:])
	return nil
}
