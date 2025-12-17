package main

import (
	"os"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/db"
	"github.com/iLeoon/realtime-gateway/internal/httpserver"
	"github.com/iLeoon/realtime-gateway/internal/router"
	"github.com/iLeoon/realtime-gateway/internal/tcp"
	"github.com/iLeoon/realtime-gateway/internal/tcpclient"
	"github.com/iLeoon/realtime-gateway/internal/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func main() {
	// A ready channel that
	tcpServerReady := make(chan struct{})
	// Load the configuration variables.
	conf, err := config.Load()
	// Start the logger.
	logger.Initlogger()

	// Connect to database.
	db, dbErr := db.Connect(conf)

	if dbErr != nil {
		logger.Error("Error on trying to connect to the database", "Error", dbErr)
		os.Exit(1)
	}

	if err != nil {
		logger.Error("can't load configuration", "Error", err)
		os.Exit(1)
	}
	// Run the TCP server.
	go tcp.InitTCPServer(conf, tcpServerReady)

	<-tcpServerReady

	//Start new WebSocket server instance.
	wsServer := websocket.NewWsServer()

	// Start new router instance and pass the WebSocket server connections map.
	router := router.NewRouter(wsServer.Clients())

	// Start a new TCP client to connect between TCP server
	// and WebSocket gateway.
	tcpClient, err := tcpclient.NewTCPClient(conf, router, wsServer.SignalToWs)

	if err != nil {
		logger.Error("Can't connect to the tcp server", "Error", err)
		os.Exit(1)
	}

	go httpserver.Start(conf, db)

	// Start the gateway.
	websocket.Start(wsServer, conf, tcpClient)

}
