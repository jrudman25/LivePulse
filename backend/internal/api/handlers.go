package api

import (
	"encoding/json"
	"net/http"
	"time"
	"strconv"

	"github.com/google/uuid"
	"github.com/jrudman25/livepulse/internal/aggregation"
	"github.com/jrudman25/livepulse/internal/events"
	"github.com/jrudman25/livepulse/internal/milestones"
	"github.com/jrudman25/livepulse/internal/storage"
)

// Server holds the API server dependencies
type Server struct {
	eventQueue *events.Queue
	aggManager *aggregation.Manager
	tracker    *milestones.Tracker
	wsHub      *WebSocketHub
	db         *storage.PostgresClient
}

// NewServer creates a new API server
func NewServer(
	eventQueue *events.Queue,
	aggManager *aggregation.Manager,
	tracker *milestones.Tracker,
	wsHub *WebSocketHub,
	db *storage.PostgresClient,
) *Server {
	return &Server{
		eventQueue: eventQueue,
		aggManager: aggManager,
		tracker:    tracker,
		wsHub:      wsHub,
		db:         db,
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"active_user_count": 0})
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

// HandleGetLiveEvents surfaces Postgres events to the Next.js frontend
func (s *Server) HandleGetLiveEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	q := r.URL.Query().Get("q")
	offsetStr := r.URL.Query().Get("offset")
	
	offset := 0
	if val, err := strconv.Atoi(offsetStr); err == nil {
		offset = val
	}
	
	eventsData, err := s.db.GetUpcomingEvents(r.Context(), 50, offset, q)
	if err != nil {
		http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}

	// Dynamically inject favorite states if a user_id is provided
	if userID != "" {
		favIDs, _ := s.db.GetUserFavorites(r.Context(), userID)
		favMap := make(map[string]bool)
		for _, fid := range favIDs {
			favMap[fid] = true
		}
		for i := range eventsData {
			if favMap[eventsData[i].ID] {
				eventsData[i].IsFavorite = true
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventsData)
}

// HandleGetEvent surfaces a single event by ID securely
func (s *Server) HandleGetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	event, err := s.db.GetEvent(r.Context(), id)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// FavoriteRequest represents the incoming JSON for favoriting
type FavoriteRequest struct {
	EventID string `json:"event_id"`
}

// HandleToggleFavorite toggles an event favorite natively on Postgres
func (s *Server) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		http.Error(w, "Unauthorized context", http.StatusUnauthorized)
		return
	}
	userID := userIDVal.(string)

	if r.Method == http.MethodGet {
		favorites, err := s.db.GetUserFavorites(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to fetch favorites", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(favorites)
		return
	}

	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EventID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		if err := s.db.AddFavorite(r.Context(), userID, req.EventID); err != nil {
			http.Error(w, "Failed to add favorite", http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodDelete {
		if err := s.db.RemoveFavorite(r.Context(), userID, req.EventID); err != nil {
			http.Error(w, "Failed to remove favorite", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
