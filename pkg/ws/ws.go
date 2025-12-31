package ws

type WsClient interface {
	SendMessage([]byte)
	GetConnectionID() uint32
}

type WsController interface {
	GetClient(uint32) (WsClient, bool)
	SignalToWs(uint32)
}
