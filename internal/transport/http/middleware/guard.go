package middleware

import (
	"net/http"
	"strings"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func AuthGuard(next http.Handler, s Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		autHeader := r.Header.Get("Authorization")
		if autHeader == "" {
			apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidAuthParameters("header", "MissingAuthHeader"))
			return
		}

		if !strings.HasPrefix(autHeader, prefix) {
			apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidAuthParameters("header", "InvalidHeaderFormat"))
			return
		}

		jwtToken := strings.TrimPrefix(autHeader, prefix)
		if jwtToken == "" {
			apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidAuthParameters("header", "MissingAuthToken"))
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
