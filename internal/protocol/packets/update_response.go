package packets

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ResponseUpdateMessagePacket represents a message delivered from the server
// back to a client.
type ResponseUpdateMessagePacket struct {
	ConversationID uint32
	MessageID      uint32
	UpdatedAt      time.Time
	ResContent     string
}

func (r *ResponseUpdateMessagePacket) String() string {
	return fmt.Sprintf("ResponseUpdateMessagePacket{ConversationID: %d, MessageID: %d, Updated_at: %v ResContent: %q}", r.ConversationID, r.MessageID, r.UpdatedAt, r.ResContent)
}

func (r *ResponseUpdateMessagePacket) Type() uint8 {
	return UpdateResponse
}

func (r *ResponseUpdateMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 12+len(r.ResContent))
	binary.BigEndian.PutUint32(b[:4], r.ConversationID)
	binary.BigEndian.PutUint32(b[4:8], r.MessageID)
	unixTime := r.UpdatedAt.UTC().Unix()

	// A check so (gosec lint) stops yelling at me for using possible out of range type casting
	if unixTime > 0xFFFFFFFF {
		unixTime = 0xFFFFFFFF
	} else if unixTime < 0 {
		unixTime = 0
	}
	binary.BigEndian.PutUint32(b[8:12], uint32(unixTime))
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

	r.UpdatedAt = time.Unix(int64(ts), 0).UTC()

	if len(b[12:]) > 512 {
		return errors.B(path, op, errors.Client, fmt.Errorf("message size(%v) hit the maximum size", len(b[12:])))
	}
	if len(b[12:]) == 0 {
		return errors.B(path, op, errors.Client, "message field is empty")
	}

	r.ResContent = string(b[12:])
	return nil
}
