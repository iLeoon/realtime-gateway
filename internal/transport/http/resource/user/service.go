package user

import (
	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Repository interface {
	GetUserById(userId string, ctx context.Context) (user *User, err error)
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(userRepo Repository) *service {
	return &service{
		repo:    userRepo,
		timeout: time.Second * 2,
	}
}

func (s *service) GetUser(userId string, ctx context.Context) (*User, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	user, err := s.repo.GetUserById(userId, ctx)
	if err != nil {
		log.Error.Println("retrieve user failed", "error", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "user")
		return nil, apiErr, statusCode
	}
	return user, nil, 0
}
