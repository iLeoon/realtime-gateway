package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

// WebSocket timing configuration.
// These help detect dead or slow connections and keep the connection alive.
// Check github.com/gorilla/websokcet chatapp example for more info.
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client represents a single WebSocket connection between the browser and the server.
//
// The WebSocket connection can only be written to from ONE goroutine.
// Because the server may need to send messages from many goroutines,
// We funnel all outgoing messages into `client.send`.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer c.conn.Close()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read websocket error:", err)
			break
		}

		fmt.Println(string(message))

	}
}
