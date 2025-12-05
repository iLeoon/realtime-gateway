package protcol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

const (
	ProtcolMagic  byte   = 0x89
	MaxPayloadLen uint32 = 1024
)

type Frame struct {
	Header  FrameHeader
	Payload packets.BuildPayload
}

type FrameHeader struct {
	Magic  uint8
	Opcode uint8
	Length uint32
}

func ConstructFrame(p packets.BuildPayload) *Frame {
	return &Frame{
		Header: FrameHeader{
			Magic:  ProtcolMagic,
			Opcode: p.Type(),
		},
		Payload: p,
	}
}

// EncodeFrame to Encode the frame in one slice and write it
// To the tcp connection
func (f *Frame) EncodeFrame(w io.Writer) error {

	err, payloadSlice := f.Payload.Encode()
	if err != nil {
		return err
	}
	sizeOfPayload := len(payloadSlice)
	if sizeOfPayload > int(MaxPayloadLen) {
		logger.Error("The payload is too large")
		return nil
	}

	//Create a buffer to hold the bytes of the exact size we need
	frame := make([]byte, 6+sizeOfPayload)

	//Encode the frame header which will always be constant number of bytes
	//It's 6 bytes in total, we create a slice in memory
	//And manually allocate each byte in the slice
	frame[0] = f.Header.Magic
	frame[1] = f.Header.Opcode
	binary.BigEndian.PutUint32(frame[2:], uint32(sizeOfPayload))
	//After we allocated the first 6 bytes as our header
	//Now we copy the payload slice into the rest of the frame
	copy(frame[6:], payloadSlice)
	//Write the frame into the tcp connection
	fmt.Println("The encoded frame: ", frame)
	w.Write(frame)
	return nil
}

func DecodeFrame(r io.Reader) (*Frame, error) {
	//Decode Frame Header first to extract the length of the payload
	//To be able to decode the payload and build the Frame
	//We know the length of the header 6 bytes, it's a constant
	//We can use that to extract the first 6 bytes which contains
	//The header data and the rest will be the payload
	header := make([]byte, 6)
	_, HeaderErr := io.ReadFull(r, header)
	if HeaderErr != nil {
		return nil, HeaderErr
	}

	magic := header[0]
	opcode := header[1]
	payloadLength := binary.BigEndian.Uint32(header[2:6])

	//We will create a slice to be able to get the actual payload
	//To contrust the packet and build the struct payload
	payload := make([]byte, payloadLength)
	pkt := packets.ConstructPacket(opcode)

	_, PayloadErr := io.ReadFull(r, payload)
	if PayloadErr != nil {
		return nil, PayloadErr
	}

	pkt.Decode(payload)

	return &Frame{
		Header: FrameHeader{
			Magic:  magic,
			Opcode: opcode,
			Length: payloadLength,
		},
		Payload: pkt,
	}, nil

}
