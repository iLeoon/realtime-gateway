package websocket

import (
	"net/http"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/session"
)

func Start(conf *config.Config, tcp session.Session) {
	ws := newWsServer()
	go ws.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		initServer(ws, w, r, tcp)
	})

	logger.Info("The websocekt server is up and running..")

	err := http.ListenAndServe(conf.Websocket.Port, nil)
	logger.Error("Error while trying to listen to the websokcet server", "Error", err)
}

func (s *wsServer) run() {
	for {
		select {
		case client := <-s.register:
			//Add the connectionID to the websocket map
			s.clients[client] = client.connectionID
			//Add the connectionID to the tcp server map
			client.transporter.OnConnect(client.connectionID)
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				//Remove the connectionID from the websocket map
				delete(s.clients, client)
				//Close the channel
				close(client.send)
			}
			//Remove the connectionID from the tcp server map
			client.transporter.DisConnect(client.connectionID)
		}
	}
}
