package websocket

import (
	"container/list"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/iLeoon/realtime-gateway/pkg/session"
)

type Server interface {
	reclaimConn(c *client)
	putConn(c *client)
	removeClient(c *client)
}

const (
	writeWait                      = 10 * time.Second
	pongWait                       = 60 * time.Second
	pingPeriod                     = (pongWait * 9) / 10
	maxMessageSize                 = 512
	clientPath     errors.PathName = "websocket/client"
)

// client represents a single WebSocket connection between the browser and the server.
//
// The WebSocket connection can only be written to from ONE goroutine.
// Because the server may need to send messages from many goroutines,
// We funnel all outgoing messages into `client.send`.
type client struct {
	userID        string
	conn          *websocket.Conn
	send          chan []byte
	server        Server
	tcpClient     session.Session
	connectionID  uint32
	burstyLimiter chan time.Time
	done          chan struct{}
	once          sync.Once
	idleElement   *list.Element
	lastActiveAt  time.Time
	isActive      bool
}

func (c *client) readPump() {
	const op errors.Op = "client.readPump"
	wsCode := websocket.CloseGoingAway
	reason := "connection closed"
	defer func() {
		c.Terminate(wsCode, reason, op)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		wrapErr := errors.B(clientPath, op, err)
		log.Error.Println("failed to read the message from the websocket", wrapErr)
		return
	}
	c.conn.SetPongHandler(func(appData string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			wrapErr := errors.B(clientPath, op, err)
			log.Error.Println("failed to read the message from the websocket", wrapErr)
		}
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				reason = "client initiated close"
				log.Info.Println("Client disconnected", "ClientID", c.connectionID)
			} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				reason = "unexpected error connection is closed"
				log.Error.Println("unexpected error shutting down websocket server", err)
			}
			return
		}

		// Add a throttle for reading messages.
		_, ok := <-c.burstyLimiter
		if !ok {
			return
		}

		c.server.reclaimConn(c)

		// Forward the messages to WriteToServer with the proper data.
		if err := c.tcpClient.WriteToServer(message); err != nil {
			log.Error.Println("faild to send message to the tcp server", err)
			if errors.Is(err, errors.Client) {
				wsCode, reason = websocket.ClosePolicyViolation, "invalid data is being used"
				return
			} else {
				wsCode, reason = websocket.CloseInternalServerErr, "unexpected error"
				return
			}
		}
		// Mark the connection inactive after reading.
		c.server.putConn(c)
	}

}

func (c *client) writePump() {
	const op errors.Op = "client.writePump"

	wsCode := websocket.CloseGoingAway
	reason := "connection is closing"
	wrapErr := errors.B(clientPath, op)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Terminate(wsCode, reason, op)
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.writeSocketMessage(message); err != nil {
				log.Error.Println(wrapErr, "failed to write the message to the socket", err)
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Error.Println(wrapErr, "failed to send the ticker", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error.Println(wrapErr, "failed to send the ticker", err)
				return
			}
		}

	}

}

func (c *client) writeSocketMessage(message []byte) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	if _, err := w.Write(message); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	// Mark the connection as inactive
	c.server.putConn(c)
	return nil

}

func (c *client) limiterFaucet() {
	ticker := time.NewTicker(200 * time.Millisecond)
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

func (c *client) Enqueue(message []byte) {
	const op errors.Op = "client.Enqueue"
	select {
	case c.send <- message:
		return

	default:
		log.Error.Println("the send channel for the sender is fulled")
		c.Terminate(websocket.ClosePolicyViolation, "unexpected error please reconnect", op)
		return
	}
}

func (c *client) Terminate(code int, reason string, op errors.Op) {
	c.once.Do(func() {
		c.server.removeClient(c)
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason)); err != nil {
			wrapErr := errors.B(clientPath, op, err)
			log.Error.Printf("failed to send close message to client %d: %v", c.connectionID, wrapErr)
		}
		c.conn.Close()
		close(c.send)
		close(c.done)
		if err := c.tcpClient.OnDisConnect(); err != nil {
			log.Error.Println("couldn't unregister this client", "ClientID", c.connectionID, "error", err)
		}
	})
}

func (c *client) ConnectionID() uint32 {
	return c.connectionID
}
