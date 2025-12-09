package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/session"
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
	conn         *websocket.Conn
	Send         chan []byte
	server       *wsServer
	tcpClient    session.Session
	ConnectionID uint32
}

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Info("Client disconnected", "ClientID", c.ConnectionID)
			}

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("Unexpected error shutting down websocket server", "Error", err)
			}
			break
		}
		// Forward the messages to ReadFromGateway with the proper data.
		readErr := c.tcpClient.ReadFromGateway(message, c.ConnectionID)
		if readErr != nil {
			logger.Error("Error on trying to read message from browser", "Error", readErr)
			break
		}
	}

}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			msg, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {

				logger.Error("An error while tring to write the message", "Error", err)
			}
			msg.Write(message)
			if err := msg.Close(); err != nil {
				logger.Error("Failed to close writer", "Error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}
