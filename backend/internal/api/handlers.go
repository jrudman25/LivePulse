package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jrudman25/livepulse/internal/aggregation"
	"github.com/jrudman25/livepulse/internal/events"
	"github.com/jrudman25/livepulse/internal/milestones"
)

// Server holds the API server dependencies
type Server struct {
	eventQueue *events.Queue
	aggManager *aggregation.Manager
	tracker    *milestones.Tracker
	wsHub      *WebSocketHub
}

// NewServer creates a new API server
func NewServer(
	eventQueue *events.Queue,
	aggManager *aggregation.Manager,
	tracker *milestones.Tracker,
	wsHub *WebSocketHub,
) *Server {
	return &Server{
		eventQueue: eventQueue,
		aggManager: aggManager,
		tracker:    tracker,
		wsHub:      wsHub,
	}
}

// CreateSessionRequest represents the request to create a session
type CreateSessionRequest struct {
	Name       string `json:"name"`
	Milestones []int  `json:"milestones,omitempty"`
}

// CreateSessionResponse represents the response when creating a session
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// HandleCreateSession creates a new session
func (s *Server) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "Untitled Event"
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Initialize milestones
	if len(req.Milestones) > 0 {
		s.tracker.InitializeSession(sessionID, req.Milestones)
	}

	// Initialize aggregation
	s.aggManager.GetOrCreateSession(sessionID)

	response := CreateSessionResponse{
		SessionID: sessionID,
		Name:      req.Name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleJoinSession allows a user to join a session
func (s *Server) HandleJoinSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	userID := r.URL.Query().Get("user_id")

	if sessionID == "" || userID == "" {
		http.Error(w, "session_id and user_id are required", http.StatusBadRequest)
		return
	}

	// Create join event
	event := events.JoinSessionEvent(sessionID, userID)

	// Enqueue event
	if !s.eventQueue.Enqueue(event) {
		http.Error(w, "Failed to enqueue event", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "joined",
		"session_id": sessionID,
		"user_id":    userID,
	})
}

// HandleGetStats returns current statistics for a session
func (s *Server) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	stats, exists := s.aggManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	snapshot := stats.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// HandleGetMilestones returns milestone progress for a session
func (s *Server) HandleGetMilestones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	milestoneList := s.tracker.GetSessionMilestones(sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": sessionID,
		"milestones": milestoneList,
	})
}

// HandleHealth is a health check endpoint
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
