package friendrequest

import "time"

type FriendRequestStatus string

const (
	StatusAccepted FriendRequestStatus = "accepted"
	StatusRejected FriendRequestStatus = "rejected"
	StatusPending  FriendRequestStatus = "pending"
)

type FriendRequest struct {
	SenderID    string              `json:"senderId"`
	RecipientID string              `json:"recipientId"`
	AuthorID    string              `json:"authorId"`
	Status      FriendRequestStatus `json:"status"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type FriendRequestBody struct {
	RecipientEmail string `json:"recipientEmail" validate:"required,email"`
}

type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Image       string `json:"displayImage"`
}

type FriendRequestDetailed struct {
	Sender    User                `json:"sender"`
	Recipient User                `json:"recipient"`
	Creator   User                `json:"creator"`
	Status    FriendRequestStatus `json:"status"`
	CreatedAt time.Time           `json:"createdAt"`
}

type FriendRequestList struct {
	Value []FriendRequestDetailed `json:"value"`
}
