package tcp

import (
	"fmt"
	"net"
	"os"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/protcol"
	"golang.org/x/text/cases"
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

		go handleConn(client)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	for {

		frame, err := protcol.DecodeFrame(conn)
		switch frame.Header.Opcode {
		case protcol.SEND_MESSAGE:
		}
		if err != nil {
			logger.Error("An Error when decoding")
		}
		fmt.Println(frame.Payload)

	}

}
