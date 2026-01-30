package conversation

import "context"

type Repository interface {
	CreateOrFindConversation(ctx context.Context, creatorId int) (err error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateConversation(ctx context.Context, creatorId string) {

}
