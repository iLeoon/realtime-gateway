package packets

type BuildPayload interface {
	Type() uint8
	Encode() (error, []byte)
	Decode([]byte) error
}
