package protocol_test

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/protocol"
	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
)

// startTestServer spins up a minimal TCP listener that mirrors the real
// server's behavior: it decodes one frame per connection and writes back
// an ErrorPacket on any client-side decode error.
func startTestServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go func(c net.Conn) {
				defer c.Close()
				c.SetDeadline(time.Now().Add(3 * time.Second))

				_, decodeErr := protocol.DecodeFrame(c)
				if decodeErr != nil && errors.Is(decodeErr, errors.Client) {
					pkt := &packets.ErrorPacket{
						Code:    errors.Client,
						Message: "invalid packet",
					}
					frame := protocol.ConstructFrame(pkt)
					frame.EncodeFrame(c)
				}
			}(conn)
		}
	}()

	return ln.Addr().String()
}

func TestCorruptFrame(t *testing.T) {
	addr := startTestServer(t)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	// Valid frame layout: [magic:1][opcode:1][length:4][payload:N]
	// Corrupt: wrong magic byte (0xFF instead of 0x89)
	frame := []byte{
		0xFF,                   // bad magic
		0x01,                   // opcode: CONNECT
		0x00, 0x00, 0x00, 0x08, // length: 8
		0x00, 0x00, 0x00, 0x01, // connectionID = 1
		0x00, 0x00, 0x00, 0x01, // userID = 1
	}
	conn.Write(frame)

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		t.Fatalf("expected ErrorPacket response, got err=%v n=%d", err, n)
	}

	raw := buf[:n]
	// Frame: [magic:1][opcode:1][length:4][code:1][message:N]
	if raw[0] != 0x89 {
		t.Fatalf("response magic mismatch: got 0x%02x", raw[0])
	}
	if raw[1] != packets.Error {
		t.Fatalf("expected ERROR opcode (%d), got %d", packets.Error, raw[1])
	}
	length := binary.BigEndian.Uint32(raw[2:6])
	code := raw[6]
	message := string(raw[7 : 6+length])
	fmt.Printf("ErrorPacket → code=%d message=%q\n", code, message)
}
