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

	// Start the logger.
	logger.Initlogger()

	// Load the configuration variables.
	conf, err := config.Load()
	if err != nil {
		logger.Error("couldn't load the configurations", "error", err)
		os.Exit(1)
	}
	// Connect to database.
	db, dbErr := db.Connect(conf)
	if dbErr != nil {
		logger.Error("error on trying to connect to the database", "error", dbErr)
		os.Exit(1)
	}

	// Run the TCP server.
	go tcp.NewServer(conf, tcpServerReady)

	<-tcpServerReady

	//Start new WebSocket server instance.
	server := websocket.New(conf)

	// Start new router instance and pass the WebSocket server connections map.
	router := router.New(server)

	// Start a new TCP Factory to manage connections between TCP server
	// and WebSocket gateway.
	tcpFactory := tcp.NewFactory(conf, router, server)

	// Retrive the handler then pass it to the http server.
	wsHandler := server.Handle(tcpFactory)

	http.Start(conf, db, wsHandler)

}
