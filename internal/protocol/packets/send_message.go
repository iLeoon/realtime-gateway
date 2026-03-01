package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// SendMessagePacket carries an outbound message from a client to another
// client or to a group.
type SendMessagePacket struct {
	ConversationID  uint32
	RecipientUserID uint32
	Content         string
}

func (s *SendMessagePacket) String() string {
	return fmt.Sprintf("SendMessagePacket{ConversationID: %d, RecipientUserID: %d, Content: %q}", s.ConversationID, s.RecipientUserID, s.Content)
}

func (s *SendMessagePacket) Type() uint8 {
	return SEND_MESSAGE
}

func (s *SendMessagePacket) Encode() ([]byte, error) {
	const path errors.PathName = "packets/send_message"
	const op errors.Op = "SendMessagePacket.Encode"
	b := make([]byte, 8+len(s.Content))

	binary.BigEndian.PutUint32(b[:4], s.ConversationID)
	binary.BigEndian.PutUint32(b[4:8], s.RecipientUserID)

	copy(b[8:], []byte(s.Content))

	return b, nil
}

func (s *SendMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/send_message"
	const op errors.Op = "SendMessagePacket.Decode"
	if len(b) < 8 {
		return errors.B(path, op, errors.Client, "send message packet length can't be less than 8")
	}

	s.ConversationID = binary.BigEndian.Uint32(b[:4])
	if s.ConversationID == 0 {
		return errors.B(path, op, "conversationID field is empty or 0")
	}

	s.RecipientUserID = binary.BigEndian.Uint32(b[4:8])
	if s.RecipientUserID == 0 {
		return errors.B(path, op, "recipientUserID field is empty or 0")
	}

	if len(b[8:]) > 512 {
		return errors.B(path, op, fmt.Errorf("message size(%v) hit the maximum size", len(b[8:])))
	}
	if len(b[8:]) == 0 {
		return errors.B(path, op, "message size can't be empty")
	}
	s.Content = string(b[8:])
	return nil
}
