package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/internal/ctx"
)

type Service interface {
	Create(ctx context.Context, creatorId string, cr ConversationRequest) (c *Conversation, a *apierror.APIError, statusCode int)
	Find(ctx context.Context, conversationId string, userId string) (c *Conversation, a *apierror.APIError, statusCode int)
	FindAll(ctx context.Context, conversationId string) (cl ConversationsList, a *apierror.APIError, statusCode int)
	GetMembers(ctx context.Context, conversationId string, userId string) (pl ParticipantsList, a *apierror.APIError, statusCode int)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	convMux := http.NewServeMux()

	convMux.HandleFunc("GET /conversations", h.List)
	convMux.HandleFunc("GET /conversations/{id}", h.GetConversation)
	convMux.HandleFunc("POST /conversations", h.Create)
	convMux.HandleFunc("GET /conversations/{id}/participants", h.GetMembers)

	return convMux
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var cr ConversationRequest

	// authenticatedId represents the conversation creator who initiated the request.
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	// Validate the json format
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidJSONFormat())
		return
	}

	// Validate the actual request body fields
	if err := validation.Validate(cr); err != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("ConversationRequest", err))
		return
	}

	conversation, apiErr, statusCode := h.service.Create(r.Context(), authenticatedId, cr)
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
