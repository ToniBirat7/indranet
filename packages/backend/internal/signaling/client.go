package signaling

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 32 * 1024 // 32KB — SDP offers can be large
)

// Client represents a single WebSocket connection in a signaling session.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	sessionID string
	role      string // "host" or "viewer"
}

// NewClient creates a client and registers it with the hub.
func NewClient(hub *Hub, conn *websocket.Conn, sessionID, role string) *Client {
	c := &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 64),
		sessionID: sessionID,
		role:      role,
	}
	hub.register <- c
	return c
}

// ReadPump reads messages from the WebSocket and relays them via the hub.
// Run in a goroutine per client.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("signaling: unexpected close", "session_id", c.sessionID, "role", c.role, "error", err)
			}
			break
		}
		// Relay to the other participant in the session
		c.hub.broadcast <- sessionMessage{sessionID: c.sessionID, payload: message}
	}
}

// WritePump writes messages from the send channel to the WebSocket.
// Run in a goroutine per client.
func (c *Client) WritePump() {
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
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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
