package tcp

import (
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
	defer func() {
		logger.Info("Tcp server connection is terminated")
		t.conn.Close()
	}()
	for {
		frame, err := protcol.DecodeFrame(t.conn)
		if err != nil {
			logger.Error("Invalid incoming data from gateway", "Error", err)
			return
		}
		switch p := frame.Payload.(type) {
		case *packets.ConnectPacket:
			t.RegisterConnectionIDs(p)
		case *packets.DisconnectPacket:
			t.UnRegisterConnectionIDs(p)
		case *packets.SendMessagePacket:
			err := t.HandleSendMessageReq(p)
			if err != nil {
				logger.Error("Error on encoding response packet", "Error", err)
				return
			}
		default:
			logger.Error("invalid incoming data from gateway of packet type: %T", p)
			return
		}

	}

}

func (t *tcpServer) HandleSendMessageReq(pkt *packets.SendMessagePacket) error {

	//Construct the response message payload
	var recipient uint32
	for id, _ := range t.clients {
		if id != pkt.ConnectionID {
			recipient = id
		}
	}

	resPkt := &packets.ResponseMessagePacket{
		ToConnectionID: recipient,
		ResContent:     pkt.Content,
	}

	frame := protcol.ConstructFrame(resPkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return err
	}
	return nil

}

func (t *tcpServer) RegisterConnectionIDs(pkt *packets.ConnectPacket) {
	t.clients[pkt.ConnectionID] = struct{}{}
}

func (t *tcpServer) UnRegisterConnectionIDs(pkt *packets.DisconnectPacket) {
	delete(t.clients, pkt.ConnectionID)
}
