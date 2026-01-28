package routes

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/conversation"
)

func Conversation(s conversation.Service) *http.ServeMux {
	convMux := http.NewServeMux()
	convMux.HandleFunc("GET /conversations/{id}", func(w http.ResponseWriter, r *http.Request) {
	})

	convMux.HandleFunc("GET /conversations", func(w http.ResponseWriter, r *http.Request) {
		conversation.List(w, r, s)
	})

	convMux.HandleFunc("POST /conversations", func(w http.ResponseWriter, r *http.Request) {
		conversation.Create(w, r, s)
	})
	return convMux
}
