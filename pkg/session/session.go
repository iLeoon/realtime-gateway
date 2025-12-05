package session

type Session interface {
	ReadFromGateway(data []byte, connectionID uint32)
	OnConnect(uint32)
	DisConnect(uint32)
}
