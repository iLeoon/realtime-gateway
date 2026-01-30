package user

import (
	"context"
)

type Repository interface {
	GetUserById(userId string, ctx context.Context) (user *User, err error)
}

type service struct {
	repo Repository
}

func NewService(userRepo Repository) *service {
	return &service{
		repo: userRepo,
	}
}

func (s *service) GetUser(userId string, ctx context.Context) (*User, error) {
	user, err := s.repo.GetUserById(userId, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil

}
