package routes

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/user"
)

func User(s user.Service) *http.ServeMux {
	userMux := http.NewServeMux()
	userMux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		user.GetUserProfile(w, r, s)
	})
	return userMux
}
