package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Service interface {
	Create(ctx context.Context, creatorId string, body ConversationRequest) (c *ConversationCreatedResponse, a *apierror.APIError, statusCode int)
	Find(ctx context.Context, conversationId string, userId string) (c *Conversation, a *apierror.APIError, statusCode int)
	FindAll(ctx context.Context, conversationId string) (cl ConversationsList, a *apierror.APIError, statusCode int)
	GetMembers(ctx context.Context, conversationId string, userId string) (pl ParticipantsList, a *apierror.APIError, statusCode int)
	UpdateParticipants(ctx context.Context, conversationId string, requesterId string, body UpdateConversationRequest) (pl ParticipantsList, a *apierror.APIError, statusCode int)
}

// Use message service
type MessageService interface {
	FindAll(ctx context.Context, conversationID string, userID string) (ml models.MessagesList, a *apierror.APIError, statusCode int)
}

type Handler struct {
	service        Service
	messageService MessageService
}

func NewHandler(s Service, ms MessageService) *Handler {
	return &Handler{
		service:        s,
		messageService: ms,
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

	// authenticatedId represents the conversation creator who initiated the request.
	authenticatedId, ok := ctx.UserId(r.Context())
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
	if err := validation.Validate(body); err != nil {
		log.Error.Println(err)
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("ConversationRequest", err))
		return
	}

	conversation, apiErr, statusCode := h.service.Create(r.Context(), authenticatedId, body)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	path := fmt.Sprintf("http://%s/api/v1.0%s", r.Host, r.URL.Path)
	w.Header().Set("Location", path+"/"+conversation.ConversationId)
	apiresponse.Send(w, http.StatusCreated, conversation)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}
	conversations, apiErr, statusCode := h.service.FindAll(r.Context(), authenticatedId)
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
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	conversationId := r.PathValue("id")
	_, err := strconv.Atoi(conversationId)
	if err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid conversation id",
			apierror.WithTarget("conversation"),
			apierror.WithInnerError("InvalidConversationIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	conversation, apiErr, statusCode := h.service.Find(r.Context(), conversationId, authenticatedId)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}
	apiresponse.Send(w, http.StatusOK, conversation)
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}
	conversationId := r.PathValue("id")
	_, err := strconv.Atoi(conversationId)
	if err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid conversation id",
			apierror.WithTarget("conversation"),
			apierror.WithInnerError("InvalidConversationIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	participants, apiErr, statusCode := h.service.GetMembers(r.Context(), conversationId, authenticatedId)
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
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	conversationId := r.PathValue("id")
	if _, err := strconv.Atoi(conversationId); err != nil {
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

	if err := validation.Validate(body); err != nil {
		log.Error.Println(err)
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("UpdateConversationRequest", err))
		return
	}

	participants, apiErr, statusCode := h.service.UpdateParticipants(r.Context(), conversationId, authenticatedId, body)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	if participants.Value == nil {
		participants.Value = []Participant{}
	}
	apiresponse.Send(w, http.StatusOK, participants)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserId(r.Context())
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
