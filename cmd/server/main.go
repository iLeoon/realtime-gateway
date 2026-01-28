package main

import (
	"os"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/db"
	"github.com/iLeoon/realtime-gateway/internal/router"
	"github.com/iLeoon/realtime-gateway/internal/transport/http"
	"github.com/iLeoon/realtime-gateway/internal/transport/tcp"
	"github.com/iLeoon/realtime-gateway/internal/transport/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func main() {
	// A ready channel that unblocks once the tcp server is up and running.
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
	go tcp.NewTcpServer(conf, tcpServerReady)

	<-tcpServerReady

	//Start new WebSocket server instance.
	wsServer := websocket.NewWsServer(conf)

	// Start new router instance and pass the WebSocket server connections map.
	router := router.NewRouter(wsServer)

	// Start a new TCP Factory to manage connections between TCP server
	// and WebSocket gateway.
	tcpFactory := tcp.NewFactory(conf, router, wsServer)

	if err != nil {
		logger.Error("Can't connect to the tcp server", "Error", err)
		os.Exit(1)
	}

	// Retrive the ws handler and pass it to the http server.
	wsHandler := wsServer.Start(tcpFactory)

	http.Start(conf, db, wsHandler)

}
