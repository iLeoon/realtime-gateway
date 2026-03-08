package user

import (
	"context"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
)

type Service interface {
	GetUser(userId string, ctx context.Context) (u *User, a *apierror.APIError, statusCode int)
	GetFriends(ctx context.Context, userID string) (FriendsList, *apierror.APIError, int)
	DeleteFriend(ctx context.Context, userID string, targetID string) (*apierror.APIError, int)
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
	userMux := http.NewServeMux()
	userMux.HandleFunc("GET /users/{id}", h.GetUserProfile)
	userMux.HandleFunc("GET /users/friends", h.GetFriends)
	userMux.HandleFunc("DELETE /users/friends/{targetId}", h.DeleteFriend)

	return userMux
}

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	targetId := r.PathValue("id")
	_, err := strconv.Atoi(targetId)
	if err != nil {
		apiErr := apierror.Build(apierror.BadRequestCode, "invalid user id", apierror.WithTarget("userId"), apierror.WithInnerError("InvalidUserIdFormatUsedInThePath"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}
	if authenticatedId != targetId {
		apiresponse.Send(w, http.StatusForbidden, apierror.UnAuthorizedUser("user"))
		return
	}

	user, apiErr, statusCode := h.service.GetUser(targetId, r.Context())
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	apiresponse.Send(w, http.StatusOK, user)
	return
}

func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	fl, apiErr, statusCode := h.service.GetFriends(r.Context(), authenticatedId)
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	if fl.Value == nil {
		fl.Value = []User{}
	}
	apiresponse.Send(w, http.StatusOK, fl)
}

func (h *Handler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
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

	if apiErr, statusCode := h.service.DeleteFriend(r.Context(), authenticatedId, targetId); apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
