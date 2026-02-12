package conversation

import (
	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

type Repository interface {
	CreateConversation(ctx context.Context, creatorId string, cr ConversationRequest) (c *Conversation, err error)
	FindConversation(ctx context.Context, conversationId string, userId string) (c *Conversation, err error)
	FindConversations(ctx context.Context, userId string) (cl ConversationsList, err error)
	FindMembers(ctx context.Context, conversationId string, userId string) (pl ParticipantsList, err error)
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

func (s *service) Create(ctx context.Context, creatorId string, cr ConversationRequest) (*Conversation, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversation, err := s.repo.CreateConversation(ctx, creatorId, cr)
	if err != nil {
		logger.Error("create conversation failed", "error", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversation")
		return nil, apiErr, statusCode
	}
	return conversation, nil, 0
}

func (s *service) Find(ctx context.Context, conversationId string, userId string) (*Conversation, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversation, err := s.repo.FindConversation(ctx, conversationId, userId)
	if err != nil {
		logger.Error("find conversation failed", "error", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversation")
		return nil, apiErr, statusCode
	}
	return conversation, nil, 0
}

func (s *service) FindAll(ctx context.Context, userId string) (ConversationsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversations, err := s.repo.FindConversations(ctx, userId)
	if err != nil {
		logger.Error("find list of conversations failed", "error", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversations")
		return conversations, apiErr, statusCode
	}
	return conversations, nil, 0
}

func (s *service) GetMembers(ctx context.Context, conversationId string, userId string) (ParticipantsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pl, err := s.repo.FindMembers(ctx, conversationId, userId)
	if err != nil {
		logger.Error("find list of conversation members failed", "error", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "participants")
		return pl, apiErr, statusCode
	}

	return pl, nil, 0
}
