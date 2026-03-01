// Package models contains shared domain types used across multiple HTTP resource packages.
// Placing types here avoids import cycles when one resource handler depends on another's types.
package models

import "time"

type Message struct {
	MessageID      string    `json:"id"`
	CreatorID      string    `json:"creatorId"`
	ConversationID string    `json:"conversationId"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MessagesList struct {
	Value []Message `json:"value"`
}

