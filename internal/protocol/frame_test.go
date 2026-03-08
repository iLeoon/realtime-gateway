package protocol_test

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"
)

const tcpAddr = "localhost:8080"

func TestCorruptFrame(t *testing.T) {
	conn, err := net.DialTimeout("tcp", tcpAddr, 2*time.Second)
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
	if raw[1] != 8 { // ERROR opcode = 8
		t.Fatalf("expected ERROR opcode (8), got %d", raw[1])
	}
	length := binary.BigEndian.Uint32(raw[2:6])
	code := raw[6]
	message := string(raw[7 : 6+length])
	fmt.Printf("ErrorPacket → code=%d message=%q\n", code, message)
}
