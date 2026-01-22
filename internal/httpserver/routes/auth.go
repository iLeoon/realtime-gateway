package routes

import (
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/token"
	"net/http"
)

func AuthRoute(service auth.Service, t token.Service) *http.ServeMux {

	authMux := http.NewServeMux()

	authMux.HandleFunc("GET /auth/login", func(w http.ResponseWriter, r *http.Request) {
		auth.LoginHandler(w, r, service)
	})

	authMux.HandleFunc("GET /auth/redirect/oauth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		auth.RedirectURLHandler(w, r, service, t)
	})

	return authMux

}
