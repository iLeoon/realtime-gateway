package protcol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/iLeoon/chatserver/pkg/protcol/errors"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

const (
	ProtcolMagic  byte   = 0x89 //The protocl magic is a constant number to verify the outgoing - incoming packet
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

	payloadSlice, err := f.Payload.Encode()
	if err != nil {
		return err
	}
	sizeOfPayload := len(payloadSlice)

	//Validate the max payload length before encoding
	if sizeOfPayload > int(MaxPayloadLen) {
		return fmt.Errorf("Size of payload: %w", errors.ErrMaxPayload)
	}

	//Create a buffer to hold the bytes of the exact size we need
	frame := make([]byte, 6+sizeOfPayload)

	//Encode the frame header which will always be constant number of bytes
	//It's 6 bytes in total, we create a slice in memory
	//And manually allocate each byte in the slice
	frame[0] = f.Header.Magic
	frame[1] = f.Header.Opcode
	binary.BigEndian.PutUint32(frame[2:], uint32(sizeOfPayload))
	f.Header.Length = uint32(sizeOfPayload)
	//After we allocated the first 6 bytes as our header
	//Now we copy the payload slice into the rest of the frame
	copy(frame[6:], payloadSlice)
	//Write the frame into the tcp connection
	w.Write(frame)
	return nil
}

func DecodeFrame(r io.Reader) (*Frame, error) {
	//DecodeFrame first extracts the length of the payload
	//To be able to decode the payload and build the Frame
	//We know the length of the header 6 bytes, it's a constant
	//We can use that to extract the first 6 bytes which contains
	//The header data and the rest will be the payload
	header := make([]byte, 6)

	//Read frame header
	n, headerErr := io.ReadFull(r, header)

	//Validate header size
	if n != 6 {
		return nil, fmt.Errorf("Unmatch: %w", errors.ErrHeaderSize)
	}

	//Check for any error while reading frame header
	if headerErr != nil {
		return nil, fmt.Errorf("%w:%v", errors.ErrReadHeader, headerErr)
	}

	//Validate the magic value from the incoming packet
	if header[0] != ProtcolMagic {
		return nil, fmt.Errorf("%w: magic value is %v", errors.ErrUnknownMagic, header[0])
	}

	magic := header[0]
	opcode := header[1]
	payloadLength := binary.BigEndian.Uint32(header[2:6])

	//Validate the payload length before decoding
	if int(payloadLength) > int(MaxPayloadLen) {
		return nil, fmt.Errorf("Payload size: %w", errors.ErrMaxPayload)
	}

	//We will create a slice to be able to get the actual payload
	//To contrust the packet and build the struct payload
	payload := make([]byte, payloadLength)
	pkt, packetErr := packets.ConstructPacket(opcode)
	if packetErr != nil {
		return nil, packetErr
	}

	//Read frame payload
	n, payloadErr := io.ReadFull(r, payload)

	//Check for any error while reading frame payload
	if payloadErr != nil {
		return nil, fmt.Errorf("%w:%v", errors.ErrReadPayload, payloadErr)
	}

	//Check if the length in the frame header matches
	//The length of the actual payload
	if n != int(payloadLength) {
		return nil, fmt.Errorf("Unmatch: %w", errors.ErrPayloadSize)
	}

	err := pkt.Decode(payload)
	if err != nil {
		return nil, err
	}

	return &Frame{
		Header: FrameHeader{
			Magic:  magic,
			Opcode: opcode,
			Length: payloadLength,
		},
		Payload: pkt,
	}, nil

}
