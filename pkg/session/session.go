package session

type Session interface {
	OnConnect(uint32)
	DisConnect(uint32)
	ReadFromGateway(data []byte, connectionID uint32)
	ReadFromServer()
}
