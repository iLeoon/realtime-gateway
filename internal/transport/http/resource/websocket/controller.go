package websocket

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

type Service interface {
	GenerateWsToken(userId string) (wsToken string, err error)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) RegsiterRoutes() *http.ServeMux {
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("GET /ws/token", h.GenerateTicket)

	return wsMux
}

func (h *Handler) GenerateTicket(w http.ResponseWriter, r *http.Request) {
	authenticatedId, ok := ctx.UserId(r.Context())
	if !ok {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.MissingUserIDContext())
		return
	}

	jwtToken, err := h.service.GenerateWsToken(authenticatedId)
	if err != nil {
		logger.Error("error on generating the websocket ticket", "error", err)
		apiresponse.Send(w, http.StatusInternalServerError, apierror.FaildToGenerateToken("GeneratingWsJwtTokenFailed"))
		return
	}

	resp := ResponseTicket{Ticket: jwtToken}
	apiresponse.Send(w, http.StatusOK, resp)
}
