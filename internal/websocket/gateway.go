package websocket

import (
	"net/http"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/session"
)

func Start(ws *wsServer, conf *config.Config, tcp session.Session) {
	go ws.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		initServer(ws, w, r, tcp)
	})

	logger.Info("The websocekt server is up and running..")

	err := http.ListenAndServe(conf.Websocket.Port, nil)
	logger.Error("Error while trying to listen to the websokcet server", "Error", err)
}
