package ws

type WsClient interface {
	SendMessage([]byte)
	GetConnectionID() uint32
}

type WsController interface {
	GetClient(string, uint32) (WsClient, bool)
	SignalToWs(SignalToWsReq)
}

type SignalToWsReq struct {
	UserID       string
	ConnectionID uint32
}
