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
	"github.com/iLeoon/realtime-gateway/pkg/log"
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
	mux.HandleFunc("PATCH /friendrequests/received/{targetID}", h.AcceptReceived)   // Accept a received request
	mux.HandleFunc("DELETE /friendrequests/sent/{targetID}", h.CancelSent)          // Cancel a sent request
	mux.HandleFunc("DELETE /friendrequests/received/{targetID}", h.DeclineReceived) // Decline a received request

	return mux
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body FriendRequestBody

	authenticatedID, ok := ctx.UserID(r.Context())
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
	errDetails, err := validation.Validate(body)
	if err != nil {
		log.Error.Println(err)
		apiresponse.Send(w, http.StatusBadRequest, apierror.Build(apierror.BadRequestCode, "failed to validate request body"))
		return
	}

	if errDetails != nil {
		apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidArgument("FriendRequest", errDetails))
		return
	}

	fr, apiErr, statusCode := h.service.Create(r.Context(), authenticatedID, body)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	path := fmt.Sprintf("%s://%s/api/v1.0%s", scheme, r.Host, r.URL.Path)
	w.Header().Set("Location", path+"/"+fr.RecipientID)
	apiresponse.Send(w, http.StatusCreated, fr)
}
func (h *Handler) Sent(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	fl, apiErr, statusCode := h.service.GetSent(r.Context(), authenticatedID)
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
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	fl, apiErr, statusCode := h.service.GetReceived(r.Context(), authenticatedID)
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
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetID := r.PathValue("targetID")
	if _, err := strconv.Atoi(targetID); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid target user id",
			apierror.WithTarget("targetID"),
			apierror.WithInnerError("InvalidTargetIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if apiErr, statusCode := h.service.AcceptReceived(r.Context(), authenticatedID, targetID); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CancelSent(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetID := r.PathValue("targetID")
	if _, err := strconv.Atoi(targetID); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid target user id",
			apierror.WithTarget("targetID"),
			apierror.WithInnerError("InvalidTargetIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if apiErr, statusCode := h.service.CancelSent(r.Context(), authenticatedID, targetID); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeclineReceived(w http.ResponseWriter, r *http.Request) {
	authenticatedID, ok := ctx.UserID(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetID := r.PathValue("targetID")
	if _, err := strconv.Atoi(targetID); err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid target user id",
			apierror.WithTarget("targetID"),
			apierror.WithInnerError("InvalidTargetIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if apiErr, statusCode := h.service.DeclineReceived(r.Context(), authenticatedID, targetID); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
