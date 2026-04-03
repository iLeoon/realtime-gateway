package tcp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/protocol"
	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
	"github.com/iLeoon/realtime-gateway/pkg/log"
)

func New() *server {
	return &server{
		db:                &noDBConn{},
		connections:       make(map[uint32]net.Conn),
		clients:           make(map[uint32][]uint32),
		userConversations: make(map[uint32]map[uint32]struct{}),
		roomManager:       make(map[uint32]map[uint32]struct{}),
	}
}

type noOpConn struct {
	net.Conn // Embed the interface so we don't have to implement 20+ methods
}

func (c *noOpConn) SetWriteDeadline(t time.Time) error { return nil }
func (c *noOpConn) Write(b []byte) (int, error)        { return len(b), nil }

type noDBConn struct {
	DBConnection
}

func (m *noDBConn) FetchMembers(userID uint32, ctx context.Context) ([]MemberShip, error) {
	return []MemberShip{{1, 10}, {2, 20}, {3, 30}, {4, 40}}, nil
}

const (
	magic = 0x89
)

var (
	ConnectPacket    = protocol.FrameHeader{Magic: magic, Opcode: 1, Length: 8}
	DisconnectPacket = protocol.FrameHeader{Magic: magic, Opcode: 2, Length: 8}
)

// BenchmarkPD for packet dispatcher function.
func BenchmarkPacketsDispatcher(b *testing.B) {
	log.SetLevel("disabled")
	b.Run("Connect", func(b *testing.B) {
		s := New()
		ctx := context.Background()
		conn := &noOpConn{}
		for i := 0; i < b.N; i++ {
			var userID uint32
			uniqueConn := uint32(i + 1)
			frame := &protocol.Frame{
				Header:  ConnectPacket,
				Payload: &packets.ConnectPacket{UserID: 1, ConnectionID: uniqueConn}}
			s.packetsDispatcher(frame, conn, &userID, ctx)
		}
	})

	b.Run("Disconnect", func(b *testing.B) {
		s := New()
		ctx := context.Background()
		conn := &noOpConn{}
		for i := 0; i < b.N; i++ {
			var userID uint32
			uniqueConn := uint32(i + 1)
			frame := &protocol.Frame{
				Header:  DisconnectPacket,
				Payload: &packets.DisconnectPacket{UserID: 1, ConnectionID: uniqueConn}}
			s.packetsDispatcher(frame, conn, &userID, ctx)
		}
	})

	b.Run("Update", func(b *testing.B) {})

}
