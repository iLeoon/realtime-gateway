package routes

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/token"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/websocket"
)

func Websocket(t token.Service) *http.ServeMux {
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("GET /ws/token", func(w http.ResponseWriter, r *http.Request) {
		websocket.GenerateTicket(w, r, t)
	})
	return wsMux
}
