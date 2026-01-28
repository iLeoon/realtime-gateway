package middleware

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/token"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func ValidateWsTicket(next http.Handler, s token.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the token is non-existent or not passed
		jwtToken := r.URL.Query().Get("token")
		if jwtToken == "" {
			apiErr := apierror.Build(apierror.UnauthorizedRequest, "User is not authenticated",
				apierror.WithTarget("token"),
				apierror.WithInnerError(apierror.InnerError{Code: "MissingWsQueryTicket"}))

			apierror.Send(w, http.StatusUnauthorized, apiErr)
			return
		}

		userId, err := s.DecodeToken(jwtToken)
		if err != nil {
			switch err.Category {
			case token.ClientError:
				logger.Error("Invalid jwt token was passed by the client", "Error", err.Err)
				apiErr := apierror.Build(apierror.UnauthorizedRequest, "User is not authenticated",
					apierror.WithTarget("token"),
					apierror.WithInnerError(apierror.InnerError{
						Code: "InvalidOrExpiredToken",
					}),
				)
				apierror.Send(w, http.StatusUnauthorized, apiErr)
				return
			case token.ServerError:
				logger.Error("Internal server error on trying to decode the jwt token", "Error", err.Err)
				apiErr := apierror.Build(apierror.InternalServerError, "Verification failed",
					apierror.WithTarget("token"),
					apierror.WithInnerError(apierror.InnerError{
						Code: "UnexpectedInternalError",
					}),
				)
				apierror.Send(w, http.StatusInternalServerError, apiErr)
				return
			default:
				logger.Error("Unexpected/Unknown error on tyring to decode the jwt token", "Error", err.Err)
				apiErr := apierror.Build(apierror.InternalServerError, "Verification failed For unkown reasons",
					apierror.WithTarget("token"),
					apierror.WithInnerError(apierror.InnerError{
						Code: "UnexpectedInternalError",
					}),
				)
				apierror.Send(w, http.StatusInternalServerError, apiErr)
				return
			}
		}

		ctx := ctx.SetUserId(r.Context(), userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
