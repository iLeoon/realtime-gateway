package middleware

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

type Service interface {
	DecodeToken(jwtToken string) (userId string, err error)
}

func ValidateWsTicket(next http.Handler, s Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the token is non-existent
		jwtToken := r.URL.Query().Get("token")
		if jwtToken == "" {
			apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidAuthParameters("QueryParameter", "MissingWsQueryTicket"))
			return
		}

		userId, err := s.DecodeToken(jwtToken)
		if err != nil {
			logger.Error("unexpected error while decoding token", "error", err)
			switch {
			case errors.Is(err, errors.Client):
				apiresponse.Send(w, http.StatusUnauthorized, apierror.InvalidToken())
				return
			case errors.Is(err, errors.Internal):
				apiresponse.Send(w, http.StatusInternalServerError, apierror.FaildToDecodeToken())
				return
			default:
				apiresponse.Send(w, http.StatusInternalServerError, apierror.FaildToDecodeToken())
				return
			}

		}

		ctx := ctx.SetUserId(r.Context(), userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
