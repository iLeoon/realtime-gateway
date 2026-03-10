package packets

const (
	Connect = iota + 1
	Disconnect
	Ping
	Pong
	SendMessage
	ResponseMessage
	UpdateMessage
	UpdateResponse
	DeleteMessage
	DeleteResponse
	Error
	Typing
	TypingResponse
	PresenceResponse
)
