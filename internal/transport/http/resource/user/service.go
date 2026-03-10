package user

import (
	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Repository interface {
	GetUserByID(userID string, ctx context.Context) (user *User, err error)
	GetFriends(ctx context.Context, userID string) (FriendsList, error)
	DeleteFriend(ctx context.Context, authenticatedID string, targetID string) error
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

func (s *service) DeleteFriend(ctx context.Context, userID string, targetID string) (*apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.repo.DeleteFriend(ctx, userID, targetID); err != nil {
		log.Error.Println("delete friend failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "friends")
		return apiErr, statusCode
	}
	return nil, 0
}

func (s *service) GetFriends(ctx context.Context, userID string) (FriendsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	fl, err := s.repo.GetFriends(ctx, userID)
	if err != nil {
		log.Error.Println("retrieve friends failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "friends")
		return fl, apiErr, statusCode
	}
	return fl, nil, 0
}

func (s *service) GetUser(userID string, ctx context.Context) (*User, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	user, err := s.repo.GetUserByID(userID, ctx)
	if err != nil {
		log.Error.Println("retrieve user failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "user")
		return nil, apiErr, statusCode
	}
	return user, nil, 0
}
