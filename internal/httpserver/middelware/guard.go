package middelware

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/jwt_"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func AuthGuard(next http.Handler, jwt jwt_.JwtInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("auth_token")
		if err != nil {
			logger.Error("authorization is non-existent in the cookie")
			http.Error(w, "not auntenticated.", http.StatusUnauthorized)
			return

		}

		userID, err := jwt.DecodeJWT(authCookie.Value)
		if err != nil {
			logger.Error("error on decoding jwt", "Error", err)
			http.Error(w, "Invalid subject", http.StatusBadRequest)
			return
		}

		ctx := ctx.SetUserIDCtx(r.Context(), userID)
		logger.Info("Attached the userID to the request", "UserID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
