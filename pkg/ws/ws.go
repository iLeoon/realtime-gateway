package ws

type WsClient interface {
	SendMessage(uint32, []byte) bool
	GetConnectionID() uint32
	Close()
}

type WsController interface {
	GetClient(uint32) WsClient
	SignalToWs(uint32)
}
