package packets

import (
	"bytes"
	"encoding/binary"

	"github.com/iLeoon/chatserver/pkg/logger"
)

type SendMessage struct {
	ConnectionID uint32
	Content      string
}

func (s *SendMessage) Type() uint8 {
	return SEND_MESSAGE
}

func (s *SendMessage) Encode() (error, []byte) {

	var body bytes.Buffer

	if err := binary.Write(&body, binary.BigEndian, s.ConnectionID); err != nil {
		logger.Error("Error encoding the connectionID into the buffer", "Error", err)
		return err, nil
	}

	//Converting the content into bytes because tcp
	//Connection only accept raw bytes
	msgToBytes := []byte(s.Content)

	//Add a layer to prevent unexpected content message size
	if len(msgToBytes) > 512 {
		logger.Error("Message content is bigger than the expected")
	}

	if len(msgToBytes) == 0 {
		logger.Error("There was no content in the playload")
	}

	body.Write(msgToBytes)

	return nil, body.Bytes()
}

func (s *SendMessage) Decode(b []byte) error {
	connetionID := binary.BigEndian.Uint32(b[:4])

	payloadContent := string(b[4:])

	s.ConnectionID = connetionID
	s.Content = payloadContent
	return nil

}
