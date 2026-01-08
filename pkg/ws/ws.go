package ws

type WsClient interface {
	SendMessage([]byte)
	GetConnectionID() uint32
}

type WsController interface {
	// GetClient(string) (WsClient, bool)
	SignalToWs(SignalToWsReq)
}

type SignalToWsReq struct {
	UserID       string
	ConnectionID uint32
}
