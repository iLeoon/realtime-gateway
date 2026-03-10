package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// AddedToConversationPacket is sent to a user when they are added to a
// conversation while already connected.
type AddedToConversationPacket struct {
	ConversationID uint32
}

func (a *AddedToConversationPacket) String() string {
	return fmt.Sprintf("AddedToConversationPacket{ConversationID: %d}", a.ConversationID)
}

func (a *AddedToConversationPacket) Type() uint8 {
	return AddedToConversation
}

func (a *AddedToConversationPacket) Encode() ([]byte, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:4], a.ConversationID)
	return b, nil
}

func (a *AddedToConversationPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/added_to_conversation"
	const op errors.Op = "AddedToConversationPacket.Decode"

	if len(b) < 4 {
		return errors.B(path, op, errors.Client, "added_to_conversation packet length can't be less than 4")
	}

	a.ConversationID = binary.BigEndian.Uint32(b[:4])
	if a.ConversationID == 0 {
		return errors.B(path, op, errors.Client, "conversationID field is empty or 0")
	}
	return nil
}
