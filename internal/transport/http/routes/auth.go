package routes

import (
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/token"
	"net/http"
)

func Auth(as auth.Service, t token.Service) *http.ServeMux {
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /auth/login", func(w http.ResponseWriter, r *http.Request) {
		auth.LoginHandler(w, r, as)
	})

	authMux.HandleFunc("GET /auth/redirect/oauth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		auth.RedirectURLHandler(w, r, as, t)
	})

	return authMux
}
