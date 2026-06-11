// Package signaling implements the WebSocket hub for WebRTC offer/answer/ICE relay.
package signaling

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Hub manages all active WebSocket connections, grouped by session ID.
// One session has two participants: host (the C++ agent) and viewer (the user client).
type Hub struct {
	sessions map[string]*Session // sessionID → Session
	mu       sync.RWMutex

	register   chan *Client
	unregister chan *Client
	broadcast  chan sessionMessage
	stopCh     chan struct{}
}

// Session holds the two WebSocket clients for a signaling session.
type Session struct {
	host   *Client
	viewer *Client
}

type sessionMessage struct {
	sessionID  string
	senderRole string // "host" | "viewer" | "" (hub-injected)
	payload    []byte
}

// NewHub creates a new signaling Hub.
func NewHub() *Hub {
	return &Hub{
		sessions:   make(map[string]*Session),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		broadcast:  make(chan sessionMessage, 256),
		stopCh:     make(chan struct{}),
	}
}

// Run is the Hub's event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case msg := <-h.broadcast:
			h.relayMessage(msg)
		case <-h.stopCh:
			return
		}
	}
}

// Stop signals the Hub's Run loop to exit.
func (h *Hub) Stop() {
	close(h.stopCh)
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, ok := h.sessions[c.sessionID]
	if !ok {
		sess = &Session{}
		h.sessions[c.sessionID] = sess
	}
	if c.role == "host" {
		sess.host = c
	} else {
		sess.viewer = c
	}
	slog.Info("signaling: client connected", "session_id", c.sessionID, "role", c.role)
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, ok := h.sessions[c.sessionID]
	if !ok {
		return
	}
	if c.role == "host" {
		sess.host = nil
	} else {
		sess.viewer = nil
	}
	// Clean up empty sessions
	if sess.host == nil && sess.viewer == nil {
		delete(h.sessions, c.sessionID)
	}
	slog.Info("signaling: client disconnected", "session_id", c.sessionID, "role", c.role)
}

func (h *Hub) relayMessage(msg sessionMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sess, ok := h.sessions[msg.sessionID]
	if !ok {
		return
	}

	send := func(c *Client) {
		if c == nil {
			return
		}
		select {
		case c.send <- msg.payload:
		default:
			slog.Warn("signaling: client send buffer full", "session_id", msg.sessionID, "role", c.role)
		}
	}

	// Relay to the opposite participant only (not the sender).
	// msg.senderRole is set by ReadPump; hub-injected messages (e.g. billing events)
	// use senderRole="" which fans out to both participants.
	switch msg.senderRole {
	case "host":
		send(sess.viewer)
	case "viewer":
		send(sess.host)
	default:
		send(sess.host)
		send(sess.viewer)
	}
}

// SendToSession sends an arbitrary message to all clients in a session.
// Called by the billing engine to deliver session_warning and session_kill events.
// Non-blocking: drops the message if the broadcast channel is full (hub stopped or overloaded).
func (h *Hub) SendToSession(sessionID string, message interface{}) {
	payload, err := json.Marshal(message)
	if err != nil {
		slog.Error("signaling: failed to marshal message", "error", err)
		return
	}
	select {
	case h.broadcast <- sessionMessage{sessionID: sessionID, payload: payload}:
	default:
		slog.Warn("signaling: broadcast channel full, dropping message", "session_id", sessionID)
	}
}
