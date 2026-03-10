package packets

import (
	"encoding/binary"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/errors"
)

// ResponsePresencePacket is fanned-out by the server to all online members
// of a user's conversations when that user comes online or goes offline.
//
// Wire format: [0:4]=UserID [4]=IsOnline(0|1)
type ResponsePresencePacket struct {
	UserID   uint32
	IsOnline bool
}

func (r *ResponsePresencePacket) String() string {
	return fmt.Sprintf("ResponsePresencePacket{UserID: %d, IsOnline: %v}", r.UserID, r.IsOnline)
}

func (r *ResponsePresencePacket) Type() uint8 {
	return PresenceResponse
}

func (r *ResponsePresencePacket) Encode() ([]byte, error) {
	b := make([]byte, 5)
	binary.BigEndian.PutUint32(b[:4], r.UserID)
	if r.IsOnline {
		b[4] = 1
	}
	return b, nil
}

func (r *ResponsePresencePacket) Decode(b []byte) error {
	const path errors.PathName = "packets/response_presence"
	const op errors.Op = "ResponsePresencePacket.Decode"

	if len(b) < 5 {
		return errors.B(path, op, errors.Client, "presence packet length can't be less than 5")
	}

	r.UserID = binary.BigEndian.Uint32(b[:4])
	if r.UserID == 0 {
		return errors.B(path, op, errors.Client, "userID field is empty or 0")
	}

	r.IsOnline = b[4] == 1
	return nil
}
