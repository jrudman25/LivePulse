package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
	"github.com/jrudman25/livepulse/internal/events"
	"github.com/jrudman25/livepulse/internal/storage"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkWSOrigin,
}

// defaultWSOrigins is the origin allowlist used when WS_ALLOWED_ORIGINS is unset.
var defaultWSOrigins = map[string]bool{
	"http://localhost:3000":           true,
	"https://livepulse-hq.vercel.app": true,
}

// checkWSOrigin enforces the WebSocket CSWSH origin allowlist. When
// WS_ALLOWED_ORIGINS (comma-separated) is set it replaces the defaults;
// clients that omit the Origin header are always allowed.
func checkWSOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if configured := parseOriginList(os.Getenv("WS_ALLOWED_ORIGINS")); len(configured) > 0 {
		return configured[origin]
	}
	return defaultWSOrigins[origin]
}

// isValidSessionID bounds session IDs to a reasonable length and charset
// (Ticketmaster IDs, UUIDs, and benchmark IDs are alphanumeric with - _).
func isValidSessionID(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

// WebSocketHub manages WebSocket connections for all sessions
type WebSocketHub struct {
	sessions map[string]*SessionHub // sessionID -> SessionHub
	mu       sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		sessions: make(map[string]*SessionHub),
	}
}

// SessionHub manages connections for a single session
type SessionHub struct {
	sessionID  string
	clients    map[*Client]bool // owned exclusively by the run() goroutine
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// NewSessionHub creates a new session hub
func NewSessionHub(sessionID string) *SessionHub {
	hub := &SessionHub{
		sessionID:  sessionID,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run()
	return hub
}

// run manages the session hub
func (h *SessionHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected to session %s (total: %d)", h.sessionID, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			log.Printf("Client disconnected from session %s (total: %d)", h.sessionID, len(h.clients))

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Client represents a WebSocket client
type Client struct {
	hub       *SessionHub
	conn      *websocket.Conn
	send      chan []byte
	sessionID string
	userID    string
}

// trySend queues an outbound message without blocking; safe against a
// send channel already closed by the hub's slow-consumer eviction.
func (c *Client) trySend(msg []byte) {
	defer func() { _ = recover() }()
	select {
	case c.send <- msg:
	default:
	}
}

// replayChatHistory pushes the most recent persisted room messages to the
// newly authenticated client so joining users can follow the conversation.
func (c *Client) replayChatHistory(redisClient *storage.RedisClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	history, err := redisClient.GetRecentChat(ctx, c.sessionID)
	if err != nil {
		log.Printf("Error loading chat history for session %s: %v", c.sessionID, err)
		return
	}

	const maxReplay = 100
	if len(history) > maxReplay {
		history = history[len(history)-maxReplay:]
	}
	for _, msg := range history {
		data, err := json.Marshal(map[string]interface{}{
			"type":    "chat",
			"message": msg,
		})
		if err != nil {
			continue
		}
		c.trySend(data)
	}
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump(eventQueue *events.Queue, redisClient *storage.RedisClient) {
	defer func() {
		if c.userID != "" { // Only safely unregister and alert if formally authenticated!
			c.hub.unregister <- c
			leaveEvent := events.LeaveSessionEvent(c.sessionID, c.userID)
			eventQueue.Enqueue(leaveEvent)
		}
		c.conn.Close()
	}()

	// Bound frame size and require authentication within 15s; the 60s
	// read deadline plus pong refresh only applies after authentication.
	c.conn.SetReadLimit(8192)
	c.conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			c.trySend([]byte(`{"type":"error","message":"Invalid JSON payload structure"}`))
			continue
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		// Handle Authentication Handshake Securely First
		if c.userID == "" {
			if msgType == "authenticate" {
				token, _ := msg["token"].(string)
				authCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				userID, err := VerifyTokenManually(authCtx, token)
				cancel()
				if err != nil {
					c.trySend([]byte(`{"type":"error","message":"Authentication invalid or expired"}`))
					break // exit pump, closing connection natively
				}

				c.userID = userID
				c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
				c.conn.SetPongHandler(func(string) error {
					c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
					return nil
				})
				c.hub.register <- c
				joinEvent := events.JoinSessionEvent(c.sessionID, c.userID)
				eventQueue.Enqueue(joinEvent)
				c.trySend([]byte(`{"type":"authenticated"}`))
				if redisClient != nil {
					go c.replayChatHistory(redisClient)
				}
				continue
			} else {
				c.trySend([]byte(`{"type":"error","message":"You must authenticate before sending events"}`))
				break // kill connection payload natively!
			}
		}

		// Handle normal routing if rigorously authenticated
		switch msgType {
		case "reaction":
			reactionType, ok := msg["reaction_type"].(string)
			if !ok {
				continue
			}
			event := events.ReactionEvent(c.sessionID, c.userID, events.ReactionType(reactionType))
			eventQueue.Enqueue(event)
		case "chat":
			text, ok := msg["text"].(string)
			if !ok {
				continue
			}
			authorName, _ := msg["author_name"].(string)

			// Simple content filter (expand this later)
			if len(text) > 500 {
				log.Printf("Chat message artificially blocked natively due to string boundaries.")
				c.trySend([]byte(`{"type":"error","message":"Message payload exceeded 500 character limit"}`))
				continue
			}

			event := events.ChatEvent(c.sessionID, c.userID, text, authorName)
			eventQueue.Enqueue(event)
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetOrCreateSessionHub gets or creates a session hub
func (h *WebSocketHub) GetOrCreateSessionHub(sessionID string) *SessionHub {
	h.mu.RLock()
	hub, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if exists {
		return hub
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock
	if hub, exists := h.sessions[sessionID]; exists {
		return hub
	}

	hub = NewSessionHub(sessionID)
	h.sessions[sessionID] = hub
	return hub
}

// BroadcastToSession broadcasts a message to all clients in a session
func (h *WebSocketHub) BroadcastToSession(sessionID string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	h.mu.RLock()
	hub, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if exists {
		select {
		case hub.broadcast <- data:
		default:
			log.Printf("Broadcast buffer full for session %s; dropping message", sessionID)
		}
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")

	if !isValidSessionID(sessionID) {
		http.Error(w, "valid session_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get or create session hub natively
	hub := s.wsHub.GetOrCreateSessionHub(sessionID)

	// Create client structurally missing UserID placeholder
	client := &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		sessionID: sessionID,
		userID:    "", // Remains blank! Authenticated intrinsically inside readPump!
	}

	// Start concurrent pumps instantly to seamlessly wait for Authentication Handshake Payload over encrypted channel
	go client.writePump() // allows server to natively kickback JSON errors organically.
	go client.readPump(s.eventQueue, s.redis)
}
