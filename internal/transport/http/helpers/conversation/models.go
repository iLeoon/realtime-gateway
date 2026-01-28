package conversation

type CreateConversationReq struct {
	RecipientId int `json:"recipientId,string" validate:"required,gt=0"`
}
