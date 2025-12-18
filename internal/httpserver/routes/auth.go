package routes

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/jwt_"
)

func AuthRoute(service auth.AuthServiceInterface, jwt jwt_.JwtInterface) *http.ServeMux {

	authMux := http.NewServeMux()

	authMux.HandleFunc("GET /auth/google/login", func(w http.ResponseWriter, r *http.Request) {
		auth.LoginHandler(w, r, service)
	})

	authMux.HandleFunc("GET /auth/redirect/oauth/google/callback/", func(w http.ResponseWriter, r *http.Request) {
		auth.RedirectURLHandler(w, r, service, jwt)
	})

	return authMux

}
