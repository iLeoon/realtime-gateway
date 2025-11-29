package tcp

import (
	"net"
	"os"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
)

type tcpServer struct {
}

func InitTCPServer(conf *config.Config) {
	conn, err := net.Listen("tcp", conf.TCPServer.Port)
	if err != nil {
		logger.Error("An error occured on creating tcp server", "Error", err)
		os.Exit(1)
	}

	defer conn.Close()
	logger.Info("TCP server is up and running")

	for {
		client, err := conn.Accept()
		if err != nil {
			logger.Error("An error occured while trying to connect a client", "Error", err)
		}

		handleConn(client)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

}
