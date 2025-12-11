package websocket

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

// Start launches the WebSocket gateway and initializes the TCP client,
// establishing the full lifecycle of the system. It sets up the gateway’s
// connection handling, starts the TCP transport, and links both components
// so messages can flow between WebSocket sessions and the TCP engine.
func Start(ws *wsServer, conf *config.Config, tcp session.Session) {
	go ws.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		initServer(ws, w, r, tcp)
	})

	logger.Info("The websocekt server is up and running..")

	err := http.ListenAndServe(conf.Websocket.WsPort, nil)
	if err != nil {
		logger.Error("Error while trying to listen to the websokcet server", "Error", err)
		return
	}
}
