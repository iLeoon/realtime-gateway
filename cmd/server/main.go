package main

import (
	"os"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/httpserver"
	"github.com/iLeoon/realtime-gateway/internal/router"
	"github.com/iLeoon/realtime-gateway/internal/tcp"
	"github.com/iLeoon/realtime-gateway/internal/tcpclient"
	"github.com/iLeoon/realtime-gateway/internal/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func main() {
	// Load the configuration variables.
	conf, err := config.Load()

	// Start the logger.
	logger.Initlogger()
	if err != nil {
		logger.Error("can't load configuration", "Error", err)
		os.Exit(1)
	}
	// Run the TCP server.
	go tcp.InitTCPServer(conf)

	//Start new WebSocket server instance.
	wsServer := websocket.NewWsServer()

	// Start new router instance and pass the WebSocket server connections map.
	router := router.NewRouter(wsServer.Clients())

	// Start a new TCP client to connect between TCP server
	// and WebSocket gateway.
	tcpClient := tcpclient.NewTCPClient(conf, router, wsServer.SignalToWs)

	go httpserver.Start(conf)

	// Start the gateway.
	websocket.Start(wsServer, conf, tcpClient)

}
