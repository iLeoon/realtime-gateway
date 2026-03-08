package packets

import (
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ErrorPacket carries a structured error back to the sender.
type ErrorPacket struct {
	Code    errors.Kind
	Message string
}

func (e *ErrorPacket) String() string {
	return fmt.Sprintf("ErrorPacket{Code: %s, Message: %q}", e.Code, e.Message)
}

func (e *ErrorPacket) Type() uint8 {
	return ERROR
}

func (e *ErrorPacket) Encode() ([]byte, error) {
	b := make([]byte, 1+len(e.Message))
	b[0] = uint8(e.Code)
	copy(b[1:], e.Message)
	return b, nil
}

func (e *ErrorPacket) Decode(b []byte) error {
	const path errors.PathName = "packets/error"
	const op errors.Op = "ErrorPacket.Decode"
	if len(b) < 1 {
		return errors.B(path, op, errors.Client, "error packet too short")
	}
	e.Code = errors.Kind(b[0])
	if e.Code == 0 {
		return errors.B(path, op, errors.Internal, "code field is empty or 0")
	}
	if len(b[1:]) > 100 {
		return errors.B(path, op, fmt.Errorf("message size(%v) hit the maximum size", len(b[1:])))
	}
	if len(b[1:]) == 0 {
		return errors.B(path, op, "message size can't be empty")
	}
	e.Message = string(b[1:])
	return nil
}
