package conversation

import (
	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Repository interface {
	CreateConversation(ctx context.Context, creatorId string, body ConversationRequest) (c *ConversationCreatedResponse, err error)
	FindConversation(ctx context.Context, conversationId string, userId string) (c *Conversation, err error)
	FindConversations(ctx context.Context, userId string) (cl ConversationsList, err error)
	FindMembers(ctx context.Context, conversationId string, userId string) (pl ParticipantsList, err error)
	UpdateParticipants(ctx context.Context, conversationId string, requesterId string, body UpdateConversationRequest) (pl ParticipantsList, err error)
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

func (s *service) Create(ctx context.Context, creatorId string, body ConversationRequest) (*ConversationCreatedResponse, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	conversation, err := s.repo.CreateConversation(ctx, creatorId, body)
	if err != nil {
		log.Error.Println("create conversation failed", err)
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
		log.Error.Println("find conversation failed", err)
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
		log.Error.Println("find list of conversations failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "conversations")
		return conversations, apiErr, statusCode
	}
	return conversations, nil, 0
}

func (s *service) UpdateParticipants(ctx context.Context, conversationId string, requesterId string, body UpdateConversationRequest) (ParticipantsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pl, err := s.repo.UpdateParticipants(ctx, conversationId, requesterId, body)
	if err != nil {
		log.Error.Println("update participants failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "participants")
		return pl, apiErr, statusCode
	}
	return pl, nil, 0
}

func (s *service) GetMembers(ctx context.Context, conversationId string, userId string) (ParticipantsList, *apierror.APIError, int) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pl, err := s.repo.FindMembers(ctx, conversationId, userId)
	if err != nil {
		log.Error.Println("find list of conversation members failed", err)
		apiErr, statusCode := apierror.ErrorMapper(err, "participants")
		return pl, apiErr, statusCode
	}

	return pl, nil, 0
}
