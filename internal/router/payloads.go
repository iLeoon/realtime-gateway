package router

type ResponseMessage struct {
	AuthorID       uint32 `json:"authorID"`
	ConversationID uint32 `json:"conversationID"`
	Content        string `json:"content"`
}
