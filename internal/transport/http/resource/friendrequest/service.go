package friendrequest

import (
	"context"
	"net/http"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Repository interface {
	Create(ctx context.Context, authorID string, body FriendRequestBody) (*FriendRequest, error)
	FindSentRequests(ctx context.Context, userID string) (FriendRequestList, error)
	FindReceivedRequests(ctx context.Context, userID string) (FriendRequestList, error)
	AcceptRequest(ctx context.Context, userID string, authorID string) error
	CancelSentRequest(ctx context.Context, authorID string, targetID string) error
	DeclineReceivedRequest(ctx context.Context, userID string, targetID string) error
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository) *service {
	return &service{
		repo:    repo,
		timeout: 2 * time.Second,
	}
}

func (s *service) Create(ctx context.Context, authorID string, body FriendRequestBody) (*FriendRequest, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	fr, err := s.repo.Create(ctx, authorID, body)
	if err != nil {
		apiErr, statusCode := apierror.ErrorMapper(err, "friendRequest")
		if statusCode == http.StatusNotFound || statusCode == http.StatusBadRequest {
			log.Info.Println("failed to create a friend request", err)
		} else {
			log.Error.Println("failed to create friend request", err)
		}
		return nil, apiErr, statusCode
	}
	return fr, nil, 0
}

func (s *service) GetSent(ctx context.Context, userID string) (FriendRequestList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	fl, err := s.repo.FindSentRequests(ctx, userID)
	if err != nil {
		log.Error.Println("find sent friend requests failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "friendRequest")
		return fl, apiErr, statusCode
	}
	return fl, nil, 0
}

func (s *service) AcceptReceived(ctx context.Context, userID string, authorID string) (*apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.repo.AcceptRequest(ctx, userID, authorID); err != nil {
		apiErr, statusCode := apierror.ErrorMapper(err, "friendRequest")
		return apiErr, statusCode
	}
	return nil, 0
}

func (s *service) CancelSent(ctx context.Context, authorID string, targetID string) (*apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.repo.CancelSentRequest(ctx, authorID, targetID); err != nil {
		apiErr, statusCode := apierror.ErrorMapper(err, "friendRequest")
		return apiErr, statusCode
	}
	return nil, 0
}

func (s *service) DeclineReceived(ctx context.Context, userID string, targetID string) (*apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.repo.DeclineReceivedRequest(ctx, userID, targetID); err != nil {
		apiErr, statusCode := apierror.ErrorMapper(err, "friendRequest")
		return apiErr, statusCode
	}
	return nil, 0
}

func (s *service) GetReceived(ctx context.Context, userID string) (FriendRequestList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	fl, err := s.repo.FindReceivedRequests(ctx, userID)
	if err != nil {
		log.Error.Println("find received friend requests failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "friendRequest")
		return fl, apiErr, statusCode
	}
	return fl, nil, 0
}
