package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ResponseTypingPacket is fanned-out by the server to all online members of
// a conversation when a user starts or stops typing.
//
// Wire format: [0:4]=ConversationID [4:8]=UserID [8]=IsTyping(0|1)
type ResponseTypingPacket struct {
	ConversationID uint32
	UserID         uint32
	IsTyping       bool
}

func (r *ResponseTypingPacket) String() string {
	return fmt.Sprintf("ResponseTypingPacket{ConversationID: %d, UserID: %d, IsTyping: %v}", r.ConversationID, r.UserID, r.IsTyping)
}

func (r *ResponseTypingPacket) Type() uint8 {
	return TYPING_RESPONSE
}

func (r *ResponseTypingPacket) Encode() ([]byte, error) {
	b := make([]byte, 9)
	binary.BigEndian.PutUint32(b[:4], r.ConversationID)
	binary.BigEndian.PutUint32(b[4:8], r.UserID)
	if r.IsTyping {
		b[8] = 1
	}
	return b, nil
}

func (r *ResponseTypingPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/response_typing"
	const op errors.Op = "ResponseTypingPacket.Decode"

	if len(b) < 9 {
		return errors.B(path, op, errors.Client, "response typing packet length can't be less than 9")
	}

	r.ConversationID = binary.BigEndian.Uint32(b[:4])
	if r.ConversationID == 0 {
		return errors.B(path, op, errors.Client, "conversationID field is empty or 0")
	}

	r.UserID = binary.BigEndian.Uint32(b[4:8])
	if r.UserID == 0 {
		return errors.B(path, op, errors.Client, "userID field is empty or 0")
	}

	r.IsTyping = b[8] == 1
	return nil
}
