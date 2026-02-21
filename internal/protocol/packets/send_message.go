package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// SendMessagePacket carries an outbound message from a client to another
// client or to a group.
type SendMessagePacket struct {
	ConnectionID uint32
	Content      string
}

func (s *SendMessagePacket) String() string {
	return fmt.Sprintf("SendMessagePacket{ConnectionID: %d, Content: %q}", s.ConnectionID, s.Content)
}

func (s *SendMessagePacket) Type() uint8 {
	return SEND_MESSAGE
}

func (s *SendMessagePacket) Encode() ([]byte, error) {
	const path errors.PathName = "packets/send_message"
	const op errors.Op = "SendMessagePacket.Encode"

	var body bytes.Buffer

	if err := binary.Write(&body, binary.BigEndian, s.ConnectionID); err != nil {
		return nil, errors.B(path, op, errors.Internal, err)
	}

	//Converte content into bytes because tcpserver only accepts raw bytes
	payloadContent := []byte(s.Content)

	body.Write(payloadContent)

	return body.Bytes(), nil
}

func (s *SendMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/send_message"
	const op errors.Op = "SendMessagePacket.Decode"
	if len(b[:4]) != 4 {
		return errors.B(path, op, "connectionID field is not 4 bytes")
	}
	s.ConnectionID = binary.BigEndian.Uint32(b[:4])
	if s.ConnectionID == 0 {
		return errors.B(path, op, "connectionID field is empty or 0")

	}

	if len(b[4:]) > 512 {
		return errors.B(path, op, fmt.Errorf("message size(%v) hit the maximum size", len(b[4:])))
	}
	payloadContent := string(b[4:])

	s.Content = payloadContent

	if len(s.Content) == 0 {
		return errors.B(path, op, "message field is empty")

	}
	return nil

}
