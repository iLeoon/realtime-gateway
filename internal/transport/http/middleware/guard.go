package middleware

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/ctx"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

func AuthGuard(next http.Handler, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidAuthParameters("cookie", "MissingAuthCookie"))
			return
		}

		jwtToken := cookie.Value
		if jwtToken == "" {
			apiresponse.Send(w, http.StatusBadRequest, apierror.InvalidAuthParameters("cookie", "MissingAuthCookie"))
			return
		}

		userID, err := s.DecodeToken(jwtToken)
		if err != nil {
			log.Error.Println("unexpected error while decoding token", err)
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

		ctx := ctx.SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
