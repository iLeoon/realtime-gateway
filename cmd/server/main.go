package main

import (
	// "github.com/iLeoon/chatserver/internal/tcpserver"
	"github.com/iLeoon/chatserver/internal/websocket"
	"github.com/iLeoon/chatserver/pkg/logger"
)

func main() {
	logger.Initlogger()
	websocket.Start()
}
