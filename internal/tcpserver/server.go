package tcpserver

import (
	"fmt"
	"net"

	"github.com/iLeoon/chatserver/internal/config"
)

type TCPServer interface {
	Run()
}

type tcpServer struct {
}

func ExeServer(conf *config.Config) {
	listen, err := net.Listen("tcp", conf.TCPServer.Port)

	fmt.Println("The server is running and waiting for connections")

	if err != nil {
		panic("Can't listen to the server")
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			panic("Can't listen to the server")
		}
		go handleconn(conn)
	}
}

func handleconn(conn net.Conn) {
	defer conn.Close()
	fmt.Println(conn)
}
