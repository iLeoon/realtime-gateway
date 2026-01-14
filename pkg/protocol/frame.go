package protocol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/iLeoon/realtime-gateway/pkg/protocol/errors"
	"github.com/iLeoon/realtime-gateway/pkg/protocol/packets"
)

const (
	protocolMagic byte   = 0x89 // Protocol identifier.
	MaxPayloadLen uint32 = 1024 // Verify payload length.
)

// Frame represents a binary protocol message exchanged between
// The WebSocket gateway and the TCP engine.
// // Structure:
//   +---------+--------------+------------
//   | Magic| Opcode| Length| Payload     |
//   +---------+--------------+------------
// The frame struct consists of
//
// Magic:   1 byte   - protocol identifier
// Opcode:  1 byte   - protocol type
// Length:  4 bytes  - payload length
// Payload: M bytes  - actual user/application data

// It encapsulates both the fixed-size frame header and the variable-length payload
type Frame struct {
	Header  FrameHeader
	Payload packets.BuildPayload
}

// FrameHeader represents the fixed-size header of every protocol frame.
// It's 6 bytes - usually contains metadata (length, type, flags, etc.)
type FrameHeader struct {
	Magic  uint8
	Opcode uint8
	Length uint32
}

// ConstructFrame initializes a new Frame using the given packet payload.
// It builds the fixed-size header and attaches the payload.
// The header's Length field is initially set to 0 and will be populated
// later by the encoder before the frame is written to the wire.
func ConstructFrame(p packets.BuildPayload) *Frame {
	return &Frame{
		Header: FrameHeader{
			Magic:  protocolMagic,
			Opcode: p.Type(),
		},
		Payload: p,
	}
}

// EncodeFrame operates on the frame method itself, to transforms
// the high-level frame struct into exactly the byte sequence expected by the protocol.
//
// It computes the payload length, writes it into the header using big-endian
// byte order, and allocates a byte slice large enough to hold the entire
// frame (6 bytes of header plus the payload). The header fields are written
// manually into the first 6 bytes, after which the payload is copied into
// the remainder of the slice. The fully encoded frame is then written to
// the underlying connection.
func (f *Frame) EncodeFrame(w io.Writer) error {
	// Get the actual payload slice which is a slice of bytes.
	payloadSlice, err := f.Payload.Encode()
	if err != nil {
		return err
	}
	// Compute length.
	sizeOfPayload := len(payloadSlice)

	// Validate the max payload length before encoding.
	if sizeOfPayload > int(MaxPayloadLen) {
		return fmt.Errorf("Size of payload: %w", errors.ErrMaxPayload)
	}

	// A slice to hold the bytes of the exact size we need.
	frame := make([]byte, 6+sizeOfPayload)

	// Allocate each byte in the slice.
	frame[0] = f.Header.Magic
	frame[1] = f.Header.Opcode
	binary.BigEndian.PutUint32(frame[2:], uint32(sizeOfPayload))
	f.Header.Length = uint32(sizeOfPayload)

	// After allocation of the first 6 bytes as our frame header
	// we copy the payload slice into the rest of the frame slice.
	copy(frame[6:], payloadSlice)

	//Write the frame into the connection
	_, writeErr := w.Write(frame)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

// DecodeFrame reads a binary frame from the underlying connection and
// reconstructs it into a Frame struct. It first reads the fixed-size
// header (6 bytes) and extracts the magic byte, packet type, and payload
// length using big-endian byte order. Once the payload length is known,
// DecodeFrame allocates a buffer of the exact size and reads the
// remaining bytes from the connection.
//
// After parsing the header and payload, DecodeFrame returns a fully
// populated Frame struct containing both the header and the payload.
func DecodeFrame(r io.Reader) (*Frame, error) {
	// We know the header length is 6 bytes.
	header := make([]byte, 6)

	//Read frame header
	n, headerErr := io.ReadFull(r, header)

	// Check if the connection is dead.
	if headerErr == io.EOF || headerErr == io.ErrUnexpectedEOF {
		return nil, io.EOF
	}

	//Validate header size
	if n != 6 {
		return nil, fmt.Errorf("Unmatch: %w", errors.ErrHeaderSize)
	}

	//Check for any error while reading frame header
	if headerErr != nil {
		return nil, fmt.Errorf("%w:%v", errors.ErrReadHeader, headerErr)
	}

	//Validate the magic value from the incoming packet
	if header[0] != protocolMagic {
		return nil, fmt.Errorf("%w: magic value is %v", errors.ErrUnknownMagic, header[0])
	}

	// Assign the frame fields
	magic := header[0]
	opcode := header[1]
	payloadLength := binary.BigEndian.Uint32(header[2:6])

	//Validate the payload length before decoding
	if int(payloadLength) > int(MaxPayloadLen) {
		return nil, fmt.Errorf("Payload size: %w", errors.ErrMaxPayload)
	}

	// Create a slice to read into the incmoing payload bytes into.
	payload := make([]byte, payloadLength)

	// Build up the frame payload based on the opcode(packet type)
	pkt, packetErr := packets.ConstructPacket(opcode)
	if packetErr != nil {
		return nil, packetErr
	}

	//Read incoming bytes into the payload slice
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

	// Decode the packet payload
	err := pkt.Decode(payload)
	if err != nil {
		return nil, err
	}

	// Return the frame
	return &Frame{
		Header: FrameHeader{
			Magic:  magic,
			Opcode: opcode,
			Length: payloadLength,
		},
		Payload: pkt,
	}, nil

}
