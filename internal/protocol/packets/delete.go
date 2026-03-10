package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

type DeleteMessagePacket struct {
	MessageID      uint32
	ConversationID uint32
}

func (d *DeleteMessagePacket) String() string {
	return fmt.Sprintf("DeleteMessagePacket{MessageID: %d, ConversationID: %d}", d.MessageID, d.ConversationID)
}
func (d *DeleteMessagePacket) Type() uint8 {
	return DeleteMessage
}

func (d *DeleteMessagePacket) Encode() ([]byte, error) {
	b := make([]byte, 8)

	binary.BigEndian.PutUint32(b[:4], d.MessageID)
	binary.BigEndian.PutUint32(b[4:8], d.ConversationID)

	return b, nil
}

func (d *DeleteMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/delet_message"
	const op errors.Op = "DeleteMessagePacket.Decode"

	if len(b) < 8 {
		return errors.B(path, op, errors.Client, "delete message packet length can't be less than 8")
	}

	d.MessageID = binary.BigEndian.Uint32(b[:4])
	if d.MessageID == 0 {
		return errors.B(path, op, "MessageID field is empty or 0")
	}
	d.ConversationID = binary.BigEndian.Uint32(b[4:8])
	if d.ConversationID == 0 {
		return errors.B(path, op, "ConversationID field is empty or 0")
	}

	return nil
}
