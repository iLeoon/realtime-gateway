package user

import (
	"context"
	"net/http"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

type Service interface {
	GetUser(userId string, ctx context.Context) (u *User, a *apierror.APIError, statusCode int)
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
		apiresponse.Send(w, http.StatusUnauthorized, apierror.UnAuthorizedUser("user"))
		return
	}

	user, apiErr, statusCode := h.service.GetUser(targetId, r.Context())
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	log.Error.Println("error while trying to fetch user", "error", err)
	apiresponse.Send(w, http.StatusOK, user)
	return
}
