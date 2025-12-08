package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/iLeoon/chatserver/pkg/protcol/errors"
)

type SendMessagePacket struct {
	ConnectionID uint32
	Content      string
}

func (s *SendMessagePacket) Type() uint8 {
	return SEND_MESSAGE
}

func (s *SendMessagePacket) Encode() ([]byte, error) {

	var body bytes.Buffer

	if err := binary.Write(&body, binary.BigEndian, s.ConnectionID); err != nil {
		return nil, err
	}

	//Converte content into bytes because tcpserver only accepts raw bytes
	payloadContent := []byte(s.Content)

	body.Write(payloadContent)

	return body.Bytes(), nil
}

func (s *SendMessagePacket) Decode(b []byte) error {
	if len(b[:4]) != 4 {
		return fmt.Errorf("%w", errors.ErrPktSize)
	}
	s.ConnectionID = binary.BigEndian.Uint32(b[:4])
	if s.ConnectionID == 0 {
		return fmt.Errorf("The connectionID cannot be 0: %w", errors.ErrPktSize)
	}

	if len(b[4:]) > 512 {
		return fmt.Errorf("Message size(%v) hit the maximum size: %w", len(b[4:]), errors.ErrPktSize)
	}
	payloadContent := string(b[4:])

	s.Content = payloadContent

	if len(s.Content) == 0 {
		return fmt.Errorf("The Message cannot be of size 0: %w", errors.ErrPktSize)
	}
	return nil

}
