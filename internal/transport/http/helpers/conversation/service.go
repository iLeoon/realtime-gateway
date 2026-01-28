package conversation

import "context"

type Service interface {
	CreateConversation(ctx context.Context, creatorId string)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateConversation(ctx context.Context, creatorId string) {

}
