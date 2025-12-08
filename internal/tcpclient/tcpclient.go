package tcpclient

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/internal/router"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/protcol"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

type ClientPayload struct {
	Opcode  string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type tcpClient struct {
	conn   net.Conn
	router *router.Router
}

func NewTCPClient(conf *config.Config, router *router.Router) *tcpClient {
	conn, err := net.Dial("tcp", conf.TCPServer.Port)
	if err != nil {
		logger.Error("Can't connect to the tcp server", "Error", err)
	}

	logger.Info("The websocket gateway successfully connected to the tcp server")

	client := &tcpClient{
		conn:   conn,
		router: router,
	}
	go client.ReadFromServer()
	return client

}

func (t *tcpClient) ReadFromGateway(data []byte, connectionID uint32) error {
	cp := &ClientPayload{}
	json.Unmarshal(data, cp)
	switch cp.Opcode {
	case "send_message":
		var data SendMessagePayload
		json.Unmarshal(cp.Payload, &data)
		pkt := &packets.SendMessagePacket{
			ConnectionID: connectionID,
			Content:      data.Content,
		}
		frame := protcol.ConstructFrame(pkt)
		err := frame.EncodeFrame(t.conn)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid packet type %s", cp.Opcode)
	}
	return nil
}

func (t *tcpClient) ReadFromServer() {
	defer func() {
		t.conn.Close()
		logger.Info("Gateway closed connection to tcp server")
	}()
	for {
		frame, err := protcol.DecodeFrame(t.conn)

		if err != nil {
			logger.Error("Invalid incoming data from tcp server", "Error", err)
			return
		}
		t.router.Route(frame)

	}
}

func (t *tcpClient) OnConnect(connectionID uint32) error {
	pkt := &packets.ConnectPacket{
		ConnectionID: connectionID,
	}

	frame := protcol.ConstructFrame(pkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return err
	}
	return nil
}

func (t *tcpClient) DisConnect(connectionID uint32) error {
	pkt := &packets.DisconnectPacket{
		ConnectionID: connectionID,
	}

	frame := protcol.ConstructFrame(pkt)
	err := frame.EncodeFrame(t.conn)
	if err != nil {
		return err
	}
	return nil
}
