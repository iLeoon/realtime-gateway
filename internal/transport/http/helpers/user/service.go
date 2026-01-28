package user

import (
	"context"
)

type Service interface {
	GetUser(userId int, ctxc context.Context) (user *User, err error)
}

type service struct {
	repo Repository
}

func NewService(userRepo Repository) Service {
	return &service{
		repo: userRepo,
	}
}

func (s *service) GetUser(userId int, ctx context.Context) (*User, error) {
	user, err := s.repo.GetUserById(userId, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil

}
