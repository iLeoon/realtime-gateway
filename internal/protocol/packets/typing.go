package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// TypingPacket is sent by a client to signal that the user has started or
// stopped typing in a conversation.
type TypingPacket struct {
	ConversationID uint32
	IsTyping       bool
}

func (t *TypingPacket) String() string {
	return fmt.Sprintf("TypingPacket{ConversationID: %d, IsTyping: %v}", t.ConversationID, t.IsTyping)
}

func (t *TypingPacket) Type() uint8 {
	return Typing
}

func (t *TypingPacket) Encode() ([]byte, error) {
	b := make([]byte, 5)
	binary.BigEndian.PutUint32(b[:4], t.ConversationID)
	if t.IsTyping {
		b[4] = 1
	}
	return b, nil
}

func (t *TypingPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/typing"
	const op errors.Op = "TypingPacket.Decode"

	if len(b) < 5 {
		return errors.B(path, op, errors.Client, "typing packet length can't be less than 5")
	}

	t.ConversationID = binary.BigEndian.Uint32(b[:4])
	if t.ConversationID == 0 {
		return errors.B(path, op, errors.Client, "conversationID field is empty or 0")
	}

	t.IsTyping = b[4] == 1
	return nil
}
