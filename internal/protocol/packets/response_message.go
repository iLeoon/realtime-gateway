package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
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
	const path errors.PathName = "packets/response_message"
	const op errors.Op = "ResponseMessagePacket.Encode"
	var body bytes.Buffer

	//Encoding each value of the slice into the buffer
	if err := binary.Write(&body, binary.BigEndian, r.ToConnectionID); err != nil {
		return nil, errors.B(path, op, errors.Internal, err)
	}
	//Converting the content into bytes because tcp
	//Connection only accept raw bytes
	payloadContent := []byte(r.ResContent)

	//Add a layer to prevent unexpected content message size
	body.Write(payloadContent)

	return body.Bytes(), nil

}

func (r *ResponseMessagePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/response_message"
	const op errors.Op = "ResponseMessagePacket.Decode"

	if len(b[:4]) != 4 {
		return errors.B(path, op, "connectionID field is not 4 bytes")
	}

	r.ToConnectionID = binary.BigEndian.Uint32(b[:4])
	if r.ToConnectionID == 0 {
		return errors.B(path, op, "connectionID field is empty or 0")
	}

	if len(b[4:]) > 512 {
		return errors.B(path, op, fmt.Errorf("message size(%v) hit the maximum size", len(b[4:])))
	}

	r.ResContent = string(b[4:])

	if len(r.ResContent) == 0 {
		return errors.B(path, op, "message field is empty")
	}
	return nil
}
