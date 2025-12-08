package session

type Session interface {
	OnConnect(uint32) error
	DisConnect(uint32) error
	ReadFromGateway(data []byte, connectionID uint32) error
	ReadFromServer()
}
