package errors

import "errors"

var (
	ErrMaxPayload   = errors.New("The payload hit the maximum size")
	ErrReadHeader   = errors.New("Error on attempting to read header from conn")
	ErrReadPayload  = errors.New("Error on attempting to read payload from conn")
	ErrPacketType   = errors.New("Unknown packet type")
	ErrUnknownMagic = errors.New("Unknown magic value")
	ErrPayloadSize  = errors.New("Header len doesn't match payload len")
	ErrHeaderSize   = errors.New("Invalid header len")

	ErrPktSize = errors.New("Invalid packet field size")
)
