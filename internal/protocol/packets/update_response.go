package packets

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ResponseMessagePacket represents a message delivered from the server
// back to a client.
type ResponseUpdateMessagePacket struct {
	ConversationID uint32
	MessageID      uint32
	Updated_at     time.Time
	ResContent     string
}

func (r *ResponseUpdateMessagePacket) String() string {
	return fmt.Sprintf("ResponseUpdateMessagePacket{ConversationID: %d, MessageID: %d, Updated_at: %v ResContent: %q}", r.ConversationID, r.MessageID, r.Updated_at, r.ResContent)
}

func (r *ResponseUpdateMessagePacket) Type() uint8 {
	return UPDATE_RESPONSE
}

func (r *ResponseUpdateMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 12+len(r.ResContent))
	binary.BigEndian.PutUint32(b[:4], r.ConversationID)
	binary.BigEndian.PutUint32(b[4:8], r.MessageID)
	binary.BigEndian.PutUint32(b[8:12], uint32(r.Updated_at.UTC().Unix()))
	copy(b[12:], r.ResContent)
	return b, nil
}

func (r *ResponseUpdateMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/update_response"
	const op errors.Op = "ResponseUpdateMessagePacket.Decode"

	if len(b) < 12 {
		return errors.B(path, op, errors.Client, "update response message packet length can't be less than 12")
	}

	r.ConversationID = binary.BigEndian.Uint32(b[:4])
	if r.ConversationID == 0 {
		return errors.B(path, op, errors.Client, "conversationID field is empty or 0")
	}

	r.MessageID = binary.BigEndian.Uint32(b[4:8])
	if r.MessageID == 0 {
		return errors.B(path, op, errors.Client, "MessageID field is empty or 0")
	}

	ts := binary.BigEndian.Uint32(b[8:12])

	r.Updated_at = time.Unix(int64(ts), 0).UTC()

	if len(b[12:]) > 512 {
		return errors.B(path, op, errors.Client, fmt.Errorf("message size(%v) hit the maximum size", len(b[12:])))
	}
	if len(b[12:]) == 0 {
		return errors.B(path, op, errors.Client, "message field is empty")
	}

	r.ResContent = string(b[12:])
	return nil
}
