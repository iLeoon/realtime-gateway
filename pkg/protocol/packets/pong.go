package packets

import "fmt"

type PongPacket struct{}

func (p *PongPacket) Type() uint8 {
	return PONG
}

func (p *PongPacket) String() string {
	return fmt.Sprintf("PongPacket")
}

func (p *PongPacket) Encode() ([]byte, error) {
	return nil, nil
}

func (p *PongPacket) Decode(b []byte) error {
	return nil
}
