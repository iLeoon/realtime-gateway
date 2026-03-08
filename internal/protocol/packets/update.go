package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

type UpdateMessagePacket struct {
	MessageID      uint32
	ConversationID uint32
	Content        string
}

func (u *UpdateMessagePacket) String() string {
	return fmt.Sprintf("UpdateMessagePacket{MessageID: %d, ConversationID: %d, Content: %s}", u.MessageID, u.ConversationID, u.Content)
}
func (u *UpdateMessagePacket) Type() uint8 {
	return UPDATE_MESSAGE
}

func (u *UpdateMessagePacket) Encode() ([]byte, error) {
	const path errors.PathName = "packets/update_message"
	const op errors.Op = "UpdateMessagePacket.Encode"
	b := make([]byte, 8+len(u.Content))

	binary.BigEndian.PutUint32(b[:4], u.MessageID)
	binary.BigEndian.PutUint32(b[4:8], u.ConversationID)

	copy(b[8:], []byte(u.Content))
	return b, nil
}

func (u *UpdateMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/update_message"
	const op errors.Op = "UpdateMessagePacket.Decode"

	if len(b) < 8 {
		return errors.B(path, op, errors.Client, "update message packet length can't be less than 8")
	}

	u.MessageID = binary.BigEndian.Uint32(b[:4])
	if u.MessageID == 0 {
		return errors.B(path, op, "connectionID field is empty or 0")
	}

	u.ConversationID = binary.BigEndian.Uint32(b[4:8])
	if u.ConversationID == 0 {
		return errors.B(path, op, "conversationID field is empty or 0")
	}

	if len(b[8:]) > 512 {
		return errors.B(path, op, fmt.Errorf("message size(%v) hit the maximum size", len(b[8:])))
	}
	if len(b[8:]) == 0 {
		return errors.B(path, op, "message size can't be empty")
	}
	u.Content = string(b[8:])
	return nil
}
