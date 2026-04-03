package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Service interface {
	Create(ctx context.Context, creatorID string, body ConversationRequest) (*ConversationCreatedResponse, *apierror.APIError, int)
	Find(ctx context.Context, conversationID string, userID string) (*Conversation, *apierror.APIError, int)
	FindAll(ctx context.Context, conversationID string) (ConversationsList, *apierror.APIError, int)
	GetMembers(ctx context.Context, conversationID string, userID string) (ParticipantsList, *apierror.APIError, int)
	UpdateParticipants(ctx context.Context, conversationID string, requesterID string, body UpdateConversationRequest) (ParticipantsList, *apierror.APIError, int)
}

type MessageService interface {
	FindAll(ctx context.Context, conversationID string, userID string) (ml models.MessagesList, a *apierror.APIError, statusCode int)
}

type Notifier interface {
	AddToRoom(userID, conversationID uint32) error
	RemoveFromRoom(userID, conversationID uint32) error
}

type Handler struct {
	service        Service
	messageService MessageService
	notifier       Notifier
}

func NewHandler(s Service, ms MessageService, notifier Notifier) *Handler {
	return &Handler{
		service:        s,
		messageService: ms,
		notifier:       notifier,
	}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	convMux := http.NewServeMux()

	convMux.HandleFunc("GET /conversations", h.List)
	convMux.HandleFunc("GET /conversations/{id}", h.GetConversation)
	convMux.HandleFunc("POST /conversations", h.Create)
	convMux.HandleFunc("GET /conversations/{id}/participants", h.GetMembers)
	convMux.HandleFunc("PATCH /conversations/{id}/participants", h.UpdateMembers)

	convMux.HandleFunc("GET /conversations/{id}/messages", h.ListMessages)

	return convMux
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body ConversationRequest

	// authenticatedID represents the conversation creator who initiated the request.
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	// Validate the json format
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidJSONFormat())
		return
	}

	// Validate the actual request body fields
	errDetails, err := validation.Validate(body)
	if err != nil {
		log.Error.Println(errors.B(errors.PathName("conversation/controller"), errors.Op("handler.Create"), err))
		apiresponse.Send(w, http.StatusBadRequest, apierror.Build(apierror.BadRequestCode, "failed to validate request body"))
		return
	}

	if errDetails != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("ConversationRequest", errDetails))
		return
	}

	conversation, apiErr, statusCode := h.service.Create(r.Context(), authenticatedID, body)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	convID, _ := strconv.ParseUint(conversation.ConversationID, 10, 32)
	creatorID, _ := strconv.ParseUint(authenticatedID, 10, 32)
	if err := h.notifier.AddToRoom(uint32(creatorID), uint32(convID)); err != nil {
		log.Error.Printf("failed to add creator %d to room %d: %v", creatorID, convID, err)
	}
	for _, id := range body.ParticipantIDs {
		if id < 0 {
			continue
		}
		if err := h.notifier.AddToRoom(uint32(id), uint32(convID)); err != nil { //nolint:gosec // participantIDs are validated positive by the request validator
			log.Error.Printf("failed to add userID %d to room %d: %v", id, convID, err)
		}
	}

	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	path := fmt.Sprintf("%s://%s/api/v1.0%s", scheme, r.Host, r.URL.Path)
	w.Header().Set("Location", path+"/"+conversation.ConversationID)
	apiresponse.Send(w, http.StatusCreated, conversation)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}
	conversations, apiErr, statusCode := h.service.FindAll(r.Context(), authenticatedID)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	// Forcing an empty array as a response if the collection is null
	if conversations.Value == nil {
		conversations.Value = []Conversation{}
	}
	apiresponse.Send(w, http.StatusOK, conversations)
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	conversationID := r.PathValue("id")
	_, err := strconv.Atoi(conversationID)
	if err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid conversation id",
			apierror.WithTarget("conversation"),
			apierror.WithInnerError("InvalidConversationIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	conversation, apiErr, statusCode := h.service.Find(r.Context(), conversationID, authenticatedID)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}
	apiresponse.Send(w, http.StatusOK, conversation)
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}
	conversationID := r.PathValue("id")
	_, err := strconv.Atoi(conversationID)
	if err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid conversation id",
			apierror.WithTarget("conversation"),
			apierror.WithInnerError("InvalidConversationIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	participants, apiErr, statusCode := h.service.GetMembers(r.Context(), conversationID, authenticatedID)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	// Forcing an empty array as a response if the collection is null
	if participants.Value == nil {
		participants.Value = []Participant{}
	}
	apiresponse.Send(w, http.StatusOK, participants)
}

func (h *Handler) UpdateMembers(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	conversationID := r.PathValue("id")
	if _, err := strconv.Atoi(conversationID); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid conversation id",
			apierror.WithTarget("conversation"),
			apierror.WithInnerError("InvalidConversationIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	var body UpdateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidJSONFormat())
		return
	}

	errDetails, err := validation.Validate(body)
	if err != nil {
		log.Error.Println(err)
		apiresponse.Send(w, http.StatusBadRequest, apierror.Build(apierror.BadRequestCode, "failed to validate request body"))
		return

	}

	if errDetails != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("UpdateConversationRequest", errDetails))
		return
	}

	participants, apiErr, statusCode := h.service.UpdateParticipants(r.Context(), conversationID, authenticatedID, body)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	convID, _ := strconv.ParseUint(conversationID, 10, 32)
	for _, id := range body.ParticipantIDs {
		if id < 0 {
			continue
		}
		if err := h.notifier.AddToRoom(uint32(id), uint32(convID)); err != nil { //nolint:gosec // participantIDs are validated positive by the request validator
			log.Error.Printf("failed to add userID %d to room %d: %v", id, convID, err)
		}
	}

	if participants.Value == nil {
		participants.Value = []Participant{}
	}
	apiresponse.Send(w, http.StatusOK, participants)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	conversationID := r.PathValue("id")
	if _, err := strconv.Atoi(conversationID); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid conversation id",
			apierror.WithTarget("conversation"),
			apierror.WithInnerError("InvalidConversationIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	ml, apiErr, statusCode := h.messageService.FindAll(r.Context(), conversationID, authenticatedID)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	if ml.Value == nil {
		ml.Value = []models.Message{}
	}
	apiresponse.Send(w, http.StatusOK, ml)
}
