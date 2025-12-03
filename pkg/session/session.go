package session

type Session interface {
	ReadFromGateway(data []byte, connectionID uint32)
}
