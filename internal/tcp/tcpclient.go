package tcp

import (
	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
	"net"
)

type tcpClient struct {
	conn *net.Conn
}

func NewTCPClient(conf *config.Config) *tcpClient {
	conn, err := net.Dial("tcp", conf.TCPServer.Port)
	if err != nil {
		logger.Error("Can't connect to the tcp server", "Error", err)
	}
	logger.Info("The websocket gateway successfully connected to the tcp server")
	return &tcpClient{
		conn: &conn,
	}

}

func (t *tcpClient) ReadFromGateway(data []byte) {
	logger.Info("The incoming Data", "Data", data)
}
