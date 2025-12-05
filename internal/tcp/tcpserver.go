package tcp

import (
	"fmt"
	"net"
	"os"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/protcol"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

type tcpServer struct {
	conn    net.Conn
	clients map[uint32]struct{}
}

func NewTcpServer(conn net.Conn) *tcpServer {
	return &tcpServer{
		clients: make(map[uint32]struct{}),
		conn:    conn,
	}
}

func InitTCPServer(conf *config.Config) {
	listner, err := net.Listen("tcp", conf.TCPServer.Port)
	if err != nil {
		logger.Error("An error occured on creating tcp server", "Error", err)
		os.Exit(1)
	}

	defer listner.Close()
	logger.Info("TCP server is up and running")

	//There will be only one connection to the tcp server
	//The tcp client which means no need for a loop at least for now
	conn, err := listner.Accept()
	if err != nil {
		logger.Error("An error occured while trying to connect a client", "Error", err)
	}
	server := NewTcpServer(conn)
	server.handleConn()
}

func (t *tcpServer) handleConn() {
	defer t.conn.Close()

	for {
		frame, err := protcol.DecodeFrame(t.conn)
		fmt.Println("The decoded frame: ", frame)
		switch p := frame.Payload.(type) {
		case *packets.SendMessage:
			t.HandleSendMessageReq(p)
		case *packets.ConnectPacket:
			t.RegisterConnectionIDs(p)
		case *packets.DisconnectPacket:
			t.UnRegisterConnectionIDs(p)
		}
		if err != nil {
			logger.Error("An Error when decoding")
		}

	}

}

func (t *tcpServer) HandleSendMessageReq(pkt *packets.SendMessage) {

	// fmt.Sprintf("Data from send message packet; ConnectionID: %d, Content: %d", pkt.ConnectionID, pkt.Content)

	//Construct the response message frame

	for index, _ := range t.clients {
		fmt.Println("The Map data", index)
	}
}

func (t *tcpServer) RegisterConnectionIDs(pkt *packets.ConnectPacket) {
	fmt.Println("Connection ID from resgiter Connection", pkt.ConnectionID)
	t.clients[pkt.ConnectionID] = struct{}{}
}

func (t *tcpServer) UnRegisterConnectionIDs(pkt *packets.DisconnectPacket) {
	fmt.Println("Connection ID to be removed from resgiter Connection", pkt.ConnectionID)
	delete(t.clients, pkt.ConnectionID)
}
