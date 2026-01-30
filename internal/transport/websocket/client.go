package websocket

import (
	"container/list"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// client represents a single WebSocket connection between the browser and the server.
//
// The WebSocket connection can only be written to from ONE goroutine.
// Because the server may need to send messages from many goroutines,
// We funnel all outgoing messages into `client.send`.
type client struct {
	userId        string
	conn          *websocket.Conn
	send          chan []byte
	server        *server
	tcpClient     session.Session
	connectionId  uint32
	burstyLimiter chan time.Time
	done          chan struct{}
	once          sync.Once
	idleElement   *list.Element
	lastActiveAt  time.Time
	isActive      bool
}

func (c *client) readPump() {
	defer func() {
		c.server.unregister <- unregisterRequest{client: c, reason: "Client side error"}
		c.server.reclaimConnLocked(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Info("Client disconnected", "ClientID", c.connectionId)
			}

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("Unexpected error shutting down websocket server", "Error", err)
			}
			return
		}

		// Add a throttle for reading messages.
		_, ok := <-c.burstyLimiter
		if !ok {
			return
		}

		c.server.reclaimConnLocked(c)

		// Forward the messages to WriteToServer with the proper data.
		readErr := c.tcpClient.WriteToServer(message)
		if readErr != nil {
			logger.Error("Error on trying to read message from browser", "Error", readErr)
			return
		}
		// Mark the connection inactive after reading.
		c.server.putConn(c)
	}

}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
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

			// Mark the connection as inactive after writing.
			c.server.putConn(c)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}

	}

}

func (c *client) limiterFaucet() {
	ticker := time.NewTicker(2 * time.Second)
	defer func() {
		ticker.Stop()
		close(c.burstyLimiter)
	}()
	for {
		select {
		case t := <-ticker.C:
			select {
			case c.burstyLimiter <- t:
			default:
			}

		case <-c.done:
			return
		}
	}
}

func (c *client) enqueue(message []byte) {
	select {
	case c.send <- message:
		return

	default:
		c.server.unregister <- unregisterRequest{client: c, reason: "Buffer full(Backpressure)"}
		return

	}
}

func (c *client) Terminate() {
	c.once.Do(func() {
		c.conn.Close()
		close(c.send)
		close(c.done)
	})
}
