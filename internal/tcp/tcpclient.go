package tcp

import (
	"encoding/json"
	"net"

	"github.com/iLeoon/chatserver/internal/config"
	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/iLeoon/chatserver/pkg/protcol"
	"github.com/iLeoon/chatserver/pkg/protcol/packets"
)

type ClientPayload struct {
	Opcode  string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type tcpClient struct {
	conn net.Conn
}

func NewTCPClient(conf *config.Config) *tcpClient {
	conn, err := net.Dial("tcp", conf.TCPServer.Port)
	if err != nil {
		logger.Error("Can't connect to the tcp server", "Error", err)
	}

	logger.Info("The websocket gateway successfully connected to the tcp server")
	return &tcpClient{
		conn: conn,
	}

}

func (t *tcpClient) ReadFromGateway(data []byte, connectionID uint32) {
	cp := &ClientPayload{}
	json.Unmarshal(data, cp)
	switch cp.Opcode {
	case "send_message":
		var data SendMessagePayload
		json.Unmarshal(cp.Payload, &data)
		pkt := &packets.SendMessage{
			ConnectionID: connectionID,
			Content:      data.Content,
		}
		frame := protcol.ConstructFrame(pkt)
		frame.EncodeFrame(t.conn)

	default:
		logger.Error("The type event is either wrong or unresgistered")

	}

}

func (t *tcpClient) OnConnect(connectionID uint32) {
	pkt := &packets.ConnectPacket{
		ConnectionID: connectionID,
	}

	frame := protcol.ConstructFrame(pkt)
	frame.EncodeFrame(t.conn)
}

func (t *tcpClient) DisConnect(connectionID uint32) {

	pkt := &packets.DisconnectPacket{
		ConnectionID: connectionID,
	}

	frame := protcol.ConstructFrame(pkt)
	frame.EncodeFrame(t.conn)
}
