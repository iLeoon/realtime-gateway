package main

import (
	// "github.com/iLeoon/chatserver/internal/tcpserver"
	"github.com/iLeoon/chatserver/internal/config"
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
	websocket.Start(conf)
}
