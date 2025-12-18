package routes

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/users"
)

func UserRoute() *http.ServeMux {

	userMux := http.NewServeMux()

	userMux.HandleFunc("GET /user/getuser", func(w http.ResponseWriter, r *http.Request) {
		users.GetUser(w, r)
	})
	return userMux
}
