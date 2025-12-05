package packets

import (
	"bytes"
	"encoding/binary"

	"github.com/iLeoon/chatserver/pkg/logger"
)

type ResponseMessagePacket struct {
	FromConnectionID uint32
	ToConnectionID   uint32
	Content          string
}

func (r *ResponseMessagePacket) Type() uint8 {
	return RESPONSE_MESSAGE
}

func (r *ResponseMessagePacket) Encode() (error, []byte) {
	var body bytes.Buffer

	if err := binary.Write(&body, binary.BigEndian, r.FromConnectionID); err != nil {
		logger.Error("Error encoding the FromconnectionID into the buffer", "Error", err)
		return err, nil
	}

	if err := binary.Write(&body, binary.BigEndian, r.ToConnectionID); err != nil {
		logger.Error("Error encoding ToconnectionID into the buffer", "Error", err)
		return err, nil
	}

	//Converting the content into bytes because tcp
	//Connection only accept raw bytes
	msgToBytes := []byte(r.Content)

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

func (r *ResponseMessagePacket) Decode(b []byte) error {
	return nil
}
