package packets

import "fmt"

type PongPacket struct{}

func (p *PongPacket) Type() uint8 {
	return Pong
}

func (p *PongPacket) String() string {
	return fmt.Sprintln("PongPacket")
}

func (p *PongPacket) Encode() ([]byte, error) {
	return nil, nil
}

func (p *PongPacket) Decode(b []byte) error {
	return nil
}
