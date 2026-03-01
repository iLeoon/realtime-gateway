package message

import (
	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Repository interface {
	FindMessages(ctx context.Context, conversationID string, userID string) (models.MessagesList, error)
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

func (s *service) FindAll(ctx context.Context, conversationID string, userID string) (models.MessagesList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ml, err := s.repo.FindMessages(ctx, conversationID, userID)
	if err != nil {
		log.Error.Println("find messages failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "messages")
		return ml, apiErr, statusCode
	}
	return ml, nil, 0
}
