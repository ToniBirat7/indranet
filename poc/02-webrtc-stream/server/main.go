// PoC 02: WebRTC Signaling Server
// Relays SDP offer/answer and ICE candidates between the C++ streamer and browser viewer.
// Sessions are identified by a room ID. Two participants (host + viewer) join the same room.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins in PoC
}

type SignalMessage struct {
	Type      string          `json:"type"`
	SDP       string          `json:"sdp,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
	Role      string          `json:"role,omitempty"` // "host" or "viewer"
}

type Room struct {
	mu      sync.Mutex
	clients map[string]*websocket.Conn // role -> conn
}

var (
	rooms   = make(map[string]*Room)
	roomsMu sync.Mutex
)

func getOrCreateRoom(id string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	if r, ok := rooms[id]; ok {
		return r
	}
	r := &Room{clients: make(map[string]*websocket.Conn)}
	rooms[id] = r
	return r
}

func handleSignal(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	role := r.URL.Query().Get("role") // "host" or "viewer"
	if roomID == "" || (role != "host" && role != "viewer") {
		http.Error(w, "missing room or role (host|viewer)", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	room := getOrCreateRoom(roomID)
	room.mu.Lock()
	room.clients[role] = conn
	log.Printf("[room %s] %s connected", roomID, role)
	room.mu.Unlock()

	defer func() {
		roomsMu.Lock()
		room.mu.Lock()
		delete(room.clients, role)
		log.Printf("[room %s] %s disconnected", roomID, role)
		if len(room.clients) == 0 {
			delete(rooms, roomID)
			log.Printf("[room %s] removed (empty)", roomID)
		}
		room.mu.Unlock()
		roomsMu.Unlock()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var signal SignalMessage
		if err := json.Unmarshal(msg, &signal); err != nil {
			log.Printf("[room %s] invalid message from %s: %v", roomID, role, err)
			continue
		}

		// Relay to the other participant
		room.mu.Lock()
		var peerRole string
		if role == "host" {
			peerRole = "viewer"
		} else {
			peerRole = "host"
		}
		peer := room.clients[peerRole]
		room.mu.Unlock()

		if peer == nil {
			log.Printf("[room %s] no %s connected yet, buffering not implemented", roomID, peerRole)
			continue
		}

		log.Printf("[room %s] relaying %s from %s to %s", roomID, signal.Type, role, peerRole)
		if err := peer.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[room %s] failed to relay to %s: %v", roomID, peerRole, err)
		}
	}
}

func main() {
	port := flag.Int("port", 8765, "Signaling server port")
	flag.Parse()

	http.HandleFunc("/signal", handleSignal)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("IndraNet PoC 02 Signaling Server listening on %s", addr)
	log.Printf("Connect: ws://localhost%s/signal?room=<id>&role=host|viewer", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
