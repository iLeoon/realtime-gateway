package conversation

import "time"

type ConversationRequest struct {
	ParticipantIDs   []int   `json:"participantIDs" validate:"required,min=1,max=50,unique,dive,gt=0"`
	ConversationType string  `json:"conversationType" validate:"required,oneof=private-chat group-chat"`
	GroupName        *string `json:"groupName" validate:"omitempty,max=50"`
}

type ConversationCreatedResponse struct {
	ConversationID   string    `json:"id"`
	CreatorID        string    `json:"creatorID"`
	ConversationType string    `json:"conversationType"`
	GroupName        *string   `json:"groupName"`
	CreatedAt        time.Time `json:"createdDate"`
}

type Participant struct {
	UserID   string    `json:"id"`
	UserName string    `json:"displayName"`
	Email    string    `json:"email"`
	Image    string    `json:"displayImage"`
	JoinedAt time.Time `json:"joinedDate"`
	Role     string    `json:"role"`
}

type Conversation struct {
	ConversationID   string        `json:"id"`
	CreatorID        string        `json:"creatorID"`
	ConversationType string        `json:"conversationType"`
	CreatedAt        time.Time     `json:"createdDate"`
	GroupName        *string       `json:"groupName"`
	Participants     []Participant `json:"participants"`
}

type UpdateConversationRequest struct {
	ParticipantIDs []int `json:"participantIDs" validate:"required,min=1,max=50,unique,dive,gt=0"`
}

type ConversationsList struct {
	Value []Conversation `json:"value"`
}

type ParticipantsList struct {
	Value []Participant `json:"value"`
}
