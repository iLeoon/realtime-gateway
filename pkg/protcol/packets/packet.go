package packets

type BuildPayload interface {
	Type() uint8
	Encode() ([]byte, error)
	Decode([]byte) error
}
