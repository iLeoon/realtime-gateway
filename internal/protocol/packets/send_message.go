package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// SendMessagePacket carries an outbound message from a client to another
// client or to a group.
type SendMessagePacket struct {
	ConversationID uint32
	Content        string
}

func (s *SendMessagePacket) String() string {
	return fmt.Sprintf("SendMessagePacket{ConversationID: %d, Content: %q}", s.ConversationID, s.Content)
}

func (s *SendMessagePacket) Type() uint8 {
	return SendMessage
}

func (s *SendMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 4+len(s.Content))

	binary.BigEndian.PutUint32(b[:4], s.ConversationID)

	copy(b[4:], []byte(s.Content))

	return b, nil
}

func (s *SendMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/send_message"
	const op errors.Op = "SendMessagePacket.Decode"
	if len(b) < 4 {
		return errors.B(path, op, errors.Client, "send message packet length can't be less than 8")
	}

	s.ConversationID = binary.BigEndian.Uint32(b[:4])
	if s.ConversationID == 0 {
		return errors.B(path, op, "conversationID field is empty or 0")
	}

	if len(b[4:]) > 512 {
		return errors.B(path, op, fmt.Errorf("message size(%v) hit the maximum size", len(b[8:])))
	}
	if len(b[4:]) == 0 {
		return errors.B(path, op, "message size can't be empty")
	}
	s.Content = string(b[4:])
	return nil
}
