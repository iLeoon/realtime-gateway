package middelware

import (
	"net/http"
	"strings"

	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/jwt_"
	"github.com/iLeoon/realtime-gateway/pkg/ctx"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func AuthGuard(next http.Handler, jwt jwt_.JwtInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			logger.Error("authorization is non-existent in the header")
			http.Error(w, "not auntenticated.", http.StatusUnauthorized)
			return
		}

		values := strings.SplitN(authHeader, " ", 2)

		if len(values) != 2 {
			logger.Error("authorization values are missing")
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}

		if values[0] != "Bearer" {
			logger.Error("authorization Bearer is missing")
			http.Error(w, "unsupported authorization scheme", http.StatusUnauthorized)
			return
		}

		userID, err := jwt.DecodeJWT(values[1])

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
