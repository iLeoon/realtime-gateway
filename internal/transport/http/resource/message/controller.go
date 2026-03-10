package message

import (
	"context"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
)

type Service interface {
	FindAll(ctx context.Context, conversationID string, userID string) (models.MessagesList, *apierror.APIError, int)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	messageMux := http.NewServeMux()
	messageMux.HandleFunc("GET /conversations/{id}/messages", h.List)
	return messageMux
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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

	ml, apiErr, statusCode := h.service.FindAll(r.Context(), conversationID, authenticatedID)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	if ml.Value == nil {
		ml.Value = []models.Message{}
	}
	apiresponse.Send(w, http.StatusOK, ml)
}
