package main

import (
	"github.com/iLeoon/chatserver/internal/config"
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
	tcpClient := tcp.NewTCPClient(conf)
	websocket.Start(conf, tcpClient)

}
