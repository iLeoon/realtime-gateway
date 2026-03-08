package router

import "time"

type ResponseMessage struct {
	AuthorID       uint32 `json:"authorID"`
	ConversationID uint32 `json:"conversationID"`
	MessageID      uint32 `json:"messageID"`
	Content        string `json:"content"`
}

type ResponseUpdateMessage struct {
	ConversationID uint32    `json:"conversationID"`
	MessageID      uint32    `json:"messageID"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Content        string    `json:"content"`
}

type ResponseDeleteMessage struct {
	MessageID      uint32 `json:"messageID"`
	ConversationID uint32 `json:"conversationID"`
	AuthorID       uint32 `json:"authorID"`
}

type ResponseTyping struct {
	ConversationID uint32 `json:"conversationID"`
	UserID         string `json:"userID"`
	IsTyping       bool   `json:"isTyping"`
}

type ResponsePresence struct {
	UserID   string `json:"userID"`
	IsOnline bool   `json:"isOnline"`
}
