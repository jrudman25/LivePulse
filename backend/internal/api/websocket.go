package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jrudman25/livepulse/internal/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, implement proper origin checking
		return true
	},
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
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
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
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected to session %s (total: %d)", h.sessionID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected from session %s (total: %d)", h.sessionID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
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

// readPump reads messages from the WebSocket connection
func (c *Client) readPump(eventQueue *events.Queue) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()

		// Send leave event
		leaveEvent := events.LeaveSessionEvent(c.sessionID, c.userID)
		eventQueue.Enqueue(leaveEvent)
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

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
			continue
		}

		// Handle different message types
		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

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
				continue // Silently drop oversized messages
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
		hub.broadcast <- data
	}
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	token := r.URL.Query().Get("token")

	if sessionID == "" || token == "" {
		http.Error(w, "session_id and token are required", http.StatusBadRequest)
		return
	}

	// Verify Clerk token
	userID, err := VerifyTokenManually(r.Context(), token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get or create session hub
	hub := s.wsHub.GetOrCreateSessionHub(sessionID)

	// Create client
	client := &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		sessionID: sessionID,
		userID:    userID,
	}

	// Register client
	hub.register <- client

	// Send join event
	joinEvent := events.JoinSessionEvent(sessionID, userID)
	s.eventQueue.Enqueue(joinEvent)

	// Start pumps
	go client.writePump()
	go client.readPump(s.eventQueue)
}
