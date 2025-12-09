package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/chatserver/pkg/protcol/errors"
)

// ResponseMessagePacket represents a message delivered from the server
// back to a client.
type ResponseMessagePacket struct {
	ToConnectionID uint32
	ResContent     string
}

func (r *ResponseMessagePacket) String() string {
	return fmt.Sprintf("ResponseMessagePacket{ToConnectionID: %d, ResContent: %q}", r.ToConnectionID, r.ResContent)
}

func (r *ResponseMessagePacket) Type() uint8 {
	return RESPONSE_MESSAGE
}

func (r *ResponseMessagePacket) Encode() ([]byte, error) {
	var body bytes.Buffer

	//Encoding each value of the slice into the buffer
	if err := binary.Write(&body, binary.BigEndian, r.ToConnectionID); err != nil {
		return nil, err
	}
	//Converting the content into bytes because tcp
	//Connection only accept raw bytes
	payloadContent := []byte(r.ResContent)

	//Add a layer to prevent unexpected content message size
	body.Write(payloadContent)

	return body.Bytes(), nil

}

func (r *ResponseMessagePacket) Decode(b []byte) error {

	if len(b[:4]) != 4 {
		return fmt.Errorf("To connectionID is not 4 bytes: %w", errors.ErrPktSize)
	}

	r.ToConnectionID = binary.BigEndian.Uint32(b[:4])
	if r.ToConnectionID == 0 {
		return fmt.Errorf("The connectionID cannot be 0: %w", errors.ErrPktSize)
	}

	if len(b[4:]) > 512 {
		return fmt.Errorf("Message size(%v) hit the maximum size: %w", len(b[4:]), errors.ErrPktSize)
	}

	r.ResContent = string(b[4:])

	if len(r.ResContent) == 0 {
		return fmt.Errorf("The Message cannot be of size 0: %w", errors.ErrPktSize)
	}
	return nil
}
