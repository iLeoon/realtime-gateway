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

	//Add a layer to prevent unexpected content message size
	if len(payloadContent) > 512 {
		return nil, fmt.Errorf("Message size(%v) hit the maximum size: %w", len(payloadContent), errors.ErrPktSize)
	}

	if len(payloadContent) == 0 {
		return nil, fmt.Errorf("Message size is 0: %w", errors.ErrPktSize)
	}

	body.Write(payloadContent)

	return body.Bytes(), nil
}

func (s *SendMessagePacket) Decode(b []byte) error {
	if len(b[:4]) != 4 {
		return fmt.Errorf("%w", errors.ErrPktSize)
	}
	s.ConnectionID = binary.BigEndian.Uint32(b[:4])

	if len(b[4:]) > 512 {
		return fmt.Errorf("Message size(%v) hit the maximum size: %w", len(b[4:]), errors.ErrPktSize)
	}
	payloadContent := string(b[4:])

	s.Content = payloadContent
	return nil

}
