package conversation

import "time"

type ConversationRequest struct {
	RecipientId      int    `json:"recipientID,string" validate:"required,gt=0"`
	ConversationType string `json:"conversationType" validate:"required,oneof=private-chat group-chat"`
	LastMessageId    *int   `json:"lastMessageId,string" validate:"omitempty,gt=0"`
}

type Participant struct {
	UserId   string    `json:"id"`
	UserName string    `json:"displayName"`
	Email    string    `json:"email"`
	JoinedAt time.Time `json:"joinedDate"`
	Role     string    `json:"role"`
}

type Conversation struct {
	ConversationId   string        `json:"id"`
	CreatorID        string        `json:"creatorId"`
	ConversationType string        `json:"conversationType"`
	LastMessageId    *int          `json:"lastMessageId"`
	CreatedAt        time.Time     `json:"createdDate"`
	Participants     []Participant `json:"participants"`
}

type ConversationsList struct {
	Value []Conversation `json:"value"`
}

type ParticipantsList struct {
	Value []Participant `json:"value"`
}
