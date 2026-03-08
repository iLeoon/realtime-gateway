package friendrequest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
)

type Service interface {
	Create(ctx context.Context, authorID string, body FriendRequestBody) (*FriendRequest, *apierror.APIError, int)
	GetSent(ctx context.Context, userID string) (FriendRequestList, *apierror.APIError, int)
	GetReceived(ctx context.Context, userID string) (FriendRequestList, *apierror.APIError, int)
	AcceptReceived(ctx context.Context, userID string, authorID string) (*apierror.APIError, int)
	CancelSent(ctx context.Context, authorID string, targetID string) (*apierror.APIError, int)
	DeclineReceived(ctx context.Context, userID string, targetID string) (*apierror.APIError, int)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /friendrequests", h.Create)                                // Create a friend request
	mux.HandleFunc("GET /friendrequests/sent", h.Sent)                              // Fetch requests sent by the author
	mux.HandleFunc("GET /friendrequests/received", h.Received)                      // Fetch requests sent to the author
	mux.HandleFunc("PATCH /friendrequests/received/{targetId}", h.AcceptReceived)   // Accept a received request
	mux.HandleFunc("DELETE /friendrequests/sent/{targetId}", h.CancelSent)          // Cancel a sent request
	mux.HandleFunc("DELETE /friendrequests/received/{targetId}", h.DeclineReceived) // Decline a received request

	return mux
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body FriendRequestBody

	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidJSONFormat())
		return
	}

	if err := validation.Validate(body); err != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("FriendRequestBody", err))
		return
	}

	fr, apiErr, statusCode := h.service.Create(r.Context(), authenticatedId, body)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	path := fmt.Sprintf("http://%s/api/v1.0%s", r.Host, r.URL.Path)
	w.Header().Set("Location", path+"/"+fr.RecipientID)
	apiresponse.Send(w, http.StatusCreated, fr)
}
func (h *Handler) Sent(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	fl, apiErr, statusCode := h.service.GetSent(r.Context(), authenticatedId)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	if fl.Value == nil {
		fl.Value = []FriendRequestDetailed{}
	}
	apiresponse.Send(w, http.StatusOK, fl)
}

func (h *Handler) Received(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	fl, apiErr, statusCode := h.service.GetReceived(r.Context(), authenticatedId)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	if fl.Value == nil {
		fl.Value = []FriendRequestDetailed{}
	}
	apiresponse.Send(w, http.StatusOK, fl)
}

func (h *Handler) AcceptReceived(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetId := r.PathValue("targetId")
	if _, err := strconv.Atoi(targetId); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid target user id",
			apierror.WithTarget("targetId"),
			apierror.WithInnerError("InvalidTargetIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if apiErr, statusCode := h.service.AcceptReceived(r.Context(), authenticatedId, targetId); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CancelSent(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetId := r.PathValue("targetId")
	if _, err := strconv.Atoi(targetId); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid target user id",
			apierror.WithTarget("targetId"),
			apierror.WithInnerError("InvalidTargetIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if apiErr, statusCode := h.service.CancelSent(r.Context(), authenticatedId, targetId); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeclineReceived(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetId := r.PathValue("targetId")
	if _, err := strconv.Atoi(targetId); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid target user id",
			apierror.WithTarget("targetId"),
			apierror.WithInnerError("InvalidTargetIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if apiErr, statusCode := h.service.DeclineReceived(r.Context(), authenticatedId, targetId); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
