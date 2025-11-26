package websocket

import (
	"net/http"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
)

func Start(conf *config.Config) {
	server := newServer()
	go server.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer(server, w, r)
	})

	logger.Info("The websocekt server is up and running..")

	err := http.ListenAndServe(conf.Websocket.Port, nil)
	logger.Error("Error while trying to listen to the websokcet server", "Error", err)
}

func (s *server) run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
		}
	}
}
