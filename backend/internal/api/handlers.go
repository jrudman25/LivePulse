package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jrudman25/livepulse/internal/aggregation"
	"github.com/jrudman25/livepulse/internal/events"
	"github.com/jrudman25/livepulse/internal/milestones"
	"github.com/jrudman25/livepulse/internal/storage"
)

// maxJSONBody bounds JSON request bodies to keep memory usage predictable.
const maxJSONBody = 1 << 20 // 1 MiB

// maxSearchQueryLen bounds the public search term forwarded to Ticketmaster.
const maxSearchQueryLen = 100

// maxMilestonesPerSession bounds milestone thresholds per created session.
const maxMilestonesPerSession = 50

// Server holds the API server dependencies
type Server struct {
	eventQueue *events.Queue
	aggManager *aggregation.Manager
	tracker    *milestones.Tracker
	wsHub      *WebSocketHub
	db         *storage.PostgresClient
	redis      *storage.RedisClient
	apiFetcher *events.APIFetcher
}

// NewServer creates a new API server
func NewServer(
	eventQueue *events.Queue,
	aggManager *aggregation.Manager,
	tracker *milestones.Tracker,
	wsHub *WebSocketHub,
	db *storage.PostgresClient,
	redis *storage.RedisClient,
	apiFetcher *events.APIFetcher,
) *Server {
	return &Server{
		eventQueue: eventQueue,
		aggManager: aggManager,
		tracker:    tracker,
		wsHub:      wsHub,
		db:         db,
		redis:      redis,
		apiFetcher: apiFetcher,
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
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "Untitled Event"
	}
	if len(req.Name) > 200 {
		http.Error(w, "name exceeds 200 characters", http.StatusBadRequest)
		return
	}
	if len(req.Milestones) > maxMilestonesPerSession {
		http.Error(w, "too many milestones (max 50)", http.StatusBadRequest)
		return
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

// HandleJoinSession allows an authenticated user to join a session.
// The user identity comes from the verified Clerk token, not request input.
func (s *Server) HandleJoinSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized context", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if !isValidSessionID(sessionID) {
		http.Error(w, "valid session_id is required", http.StatusBadRequest)
		return
	}

	// The session must already exist (created via the API, an active
	// WebSocket room, or a known event) so joins cannot mint arbitrary state.
	if _, exists := s.aggManager.GetSession(sessionID); !exists {
		if _, err := s.db.GetEvent(r.Context(), sessionID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to validate session", http.StatusInternalServerError)
			return
		}
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

	// Batch mode: session_ids=a,b,c returns a {session_id: active_user_count} map
	// so listing pages can fetch all visible room counts in one request.
	if sessionIDsParam := r.URL.Query().Get("session_ids"); sessionIDsParam != "" {
		ids := strings.Split(sessionIDsParam, ",")
		if len(ids) > 200 {
			http.Error(w, "too many session_ids (max 200)", http.StatusBadRequest)
			return
		}
		counts := make(map[string]int, len(ids))
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if stats, exists := s.aggManager.GetSession(id); exists {
				counts[id] = stats.GetActiveUserCount()
			} else {
				counts[id] = 0
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(counts)
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

// HandleHealth is a lightweight liveness check
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleReady reports whether the dependencies the service needs are
// reachable. It is intended for deployment readiness probes.
func (s *Server) HandleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	status := map[string]string{}
	ready := true

	if err := s.db.Ping(ctx); err != nil {
		status["postgres"] = err.Error()
		ready = false
	} else {
		status["postgres"] = "ok"
	}

	if s.redis == nil {
		status["redis"] = "not configured"
		ready = false
	} else if err := s.redis.Ping(ctx); err != nil {
		status["redis"] = err.Error()
		ready = false
	} else {
		status["redis"] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	if !ready {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": map[bool]string{true: "ready", false: "not ready"}[ready],
		"checks": status,
	})
}

// HandleGetLiveEvents surfaces Postgres events to the Next.js frontend
func (s *Server) HandleGetLiveEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Favorite decoration uses the verified Clerk identity injected by
	// OptionalClerkMiddleware; client-supplied user IDs are never trusted.
	userID, _ := UserIDFromContext(r.Context())

	q := r.URL.Query().Get("q")
	if len(q) > maxSearchQueryLen {
		q = q[:maxSearchQueryLen]
	}
	offsetStr := r.URL.Query().Get("offset")

	offset := 0
	if val, err := strconv.Atoi(offsetStr); err == nil && val > 0 {
		offset = val
	}

	// Infinite Search Interceptor: Fire to TM specifically if query isn't empty, gracefully load DB implicitly!
	if q != "" {
		s.apiFetcher.FetchSearchKeyword(r.Context(), q)
	}

	eventsData, err := s.db.GetUpcomingEvents(r.Context(), 200, offset, q)
	if err != nil {
		http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}

	// Dynamically inject favorite states when the request is authenticated
	if userID != "" {
		favIDs, err := s.db.GetUserFavorites(r.Context(), userID)
		if err != nil {
			log.Printf("Failed to load favorites for decorated listing: %v", err)
			http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
			return
		}
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
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to fetch event %s: %v", id, err)
		http.Error(w, "Failed to retrieve event", http.StatusInternalServerError)
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

	userID, ok := UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized context", http.StatusUnauthorized)
		return
	}

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
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
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
