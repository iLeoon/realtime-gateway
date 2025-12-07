package main

import (
	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/internal/router"
	"github.com/iLeoon/chatserver/internal/tcp"
	"github.com/iLeoon/chatserver/internal/websocket"
	"github.com/iLeoon/chatserver/pkg/logger"
)

func main() {
	conf, err := config.Load()
	logger.Initlogger()

	if err != nil {
		logger.Error("can't load configuration", "Error", err)
		panic(err)
	}
	go tcp.InitTCPServer(conf)
	wsServer := websocket.NewWsServer()
	router := router.NewRouter(wsServer.Clients())
	tcpClient := tcp.NewTCPClient(conf, router)
	websocket.Start(wsServer, conf, tcpClient)

}
