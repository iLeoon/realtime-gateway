package packets

// BuildPayload defines the common behavior shared by all protocol packet
// types. Each concrete packet embeds or implements this interface to
// provide a unified way to encode itself into bytes, decode bytes back
// into the packet, and expose its packet type.
//
// The interface serves as an abstraction allowing higher-level components to operate on packets generically
// without needing to know their concrete struct types.
// All packets in the protocol implement BuildPayload, allowing the system
// to construct, encode, decode, and dispatch packets in a consistent and
// extensible manner.

type BuildPayload interface {
	Type() uint8             // Returns the packet’s opcode/type identifier.
	Encode() ([]byte, error) // Serializes the packet’s fields into a payload byte slice.
	Decode([]byte) error     // Populates the packet’s fields by parsing the provided payload.
	String() string          // Returns a human-readable representation of the packet for logging and debugging
}
