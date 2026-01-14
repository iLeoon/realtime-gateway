package packets

import "fmt"

type PingPacket struct{}

func (p *PingPacket) Type() uint8 {
	return PING
}

func (p *PingPacket) String() string {
	return fmt.Sprintf("PingPacket")
}

func (p *PingPacket) Encode() ([]byte, error) {
	return nil, nil
}

func (p *PingPacket) Decode(b []byte) error {
	return nil
}
