package conversation

import (
	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Repository interface {
	CreateConversation(ctx context.Context, creatorID string, body ConversationRequest) (c *ConversationCreatedResponse, err error)
	FindConversation(ctx context.Context, conversationID string, userID string) (c *Conversation, err error)
	FindConversations(ctx context.Context, userID string) (cl ConversationsList, err error)
	FindMembers(ctx context.Context, conversationID string, userID string) (pl ParticipantsList, err error)
	UpdateParticipants(ctx context.Context, conversationID string, requesterID string, body UpdateConversationRequest) (pl ParticipantsList, err error)
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

func (s *service) Create(ctx context.Context, creatorID string, body ConversationRequest) (*ConversationCreatedResponse, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversation, err := s.repo.CreateConversation(ctx, creatorID, body)
	if err != nil {
		log.Error.Println("create conversation failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversation")
		return nil, apiErr, statusCode
	}
	return conversation, nil, 0
}

func (s *service) Find(ctx context.Context, conversationID string, userID string) (*Conversation, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversation, err := s.repo.FindConversation(ctx, conversationID, userID)
	if err != nil {
		log.Error.Println("find conversation failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversation")
		return nil, apiErr, statusCode
	}
	return conversation, nil, 0
}

func (s *service) FindAll(ctx context.Context, userID string) (ConversationsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversations, err := s.repo.FindConversations(ctx, userID)
	if err != nil {
		log.Error.Println("find list of conversations failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversations")
		return conversations, apiErr, statusCode
	}
	return conversations, nil, 0
}

func (s *service) UpdateParticipants(ctx context.Context, conversationID string, requesterID string, body UpdateConversationRequest) (ParticipantsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pl, err := s.repo.UpdateParticipants(ctx, conversationID, requesterID, body)
	if err != nil {
		log.Error.Println("update participants failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "participants")
		return pl, apiErr, statusCode
	}
	return pl, nil, 0
}

func (s *service) GetMembers(ctx context.Context, conversationID string, userID string) (ParticipantsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pl, err := s.repo.FindMembers(ctx, conversationID, userID)
	if err != nil {
		log.Error.Println("find list of conversation members failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "participants")
		return pl, apiErr, statusCode
	}

	return pl, nil, 0
}
