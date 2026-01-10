package aggregation

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/jrudman25/livepulse/internal/events"
)

// SessionStats holds real-time statistics for a session
type SessionStats struct {
	SessionID         string
	ActiveUsers       map[string]bool // UserID -> active status
	ReactionCounts    map[events.ReactionType]*int64
	TotalReactions    *int64
	PeakConcurrentUsers int
	StartTime         time.Time
	LastActivity      time.Time
	mu                sync.RWMutex
}

// NewSessionStats creates a new session statistics tracker
func NewSessionStats(sessionID string) *SessionStats {
	totalReactions := int64(0)
	
	return &SessionStats{
		SessionID:      sessionID,
		ActiveUsers:    make(map[string]bool),
		ReactionCounts: map[events.ReactionType]*int64{
			events.ReactionLike:     new(int64),
			events.ReactionLove:     new(int64),
			events.ReactionCheer:    new(int64),
			events.ReactionApplause: new(int64),
			events.ReactionFire:     new(int64),
			events.ReactionHeart:    new(int64),
		},
		TotalReactions:      &totalReactions,
		PeakConcurrentUsers: 0,
		StartTime:           time.Now().UTC(),
		LastActivity:        time.Now().UTC(),
	}
}

// AddUser adds a user to the active users set
func (s *SessionStats) AddUser(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ActiveUsers[userID] = true
	s.LastActivity = time.Now().UTC()
	
	currentCount := len(s.ActiveUsers)
	if currentCount > s.PeakConcurrentUsers {
		s.PeakConcurrentUsers = currentCount
	}
	
	return currentCount
}

// RemoveUser removes a user from the active users set
func (s *SessionStats) RemoveUser(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.ActiveUsers, userID)
	s.LastActivity = time.Now().UTC()
	
	return len(s.ActiveUsers)
}

// IncrementReaction atomically increments the count for a reaction type
func (s *SessionStats) IncrementReaction(reactionType events.ReactionType) int64 {
	s.mu.Lock()
	counter, exists := s.ReactionCounts[reactionType]
	s.mu.Unlock()

	if !exists {
		// Unknown reaction type, just increment total
		return atomic.AddInt64(s.TotalReactions, 1)
	}

	// Atomically increment both the specific reaction and total
	atomic.AddInt64(counter, 1)
	total := atomic.AddInt64(s.TotalReactions, 1)

	s.mu.Lock()
	s.LastActivity = time.Now().UTC()
	s.mu.Unlock()

	return total
}

// GetActiveUserCount returns the current number of active users
func (s *SessionStats) GetActiveUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.ActiveUsers)
}

// GetTotalReactions returns the total number of reactions
func (s *SessionStats) GetTotalReactions() int64 {
	return atomic.LoadInt64(s.TotalReactions)
}

// GetReactionCount returns the count for a specific reaction type
func (s *SessionStats) GetReactionCount(reactionType events.ReactionType) int64 {
	s.mu.RLock()
	counter, exists := s.ReactionCounts[reactionType]
	s.mu.RUnlock()

	if !exists {
		return 0
	}
	return atomic.LoadInt64(counter)
}

// GetAllReactionCounts returns a snapshot of all reaction counts
func (s *SessionStats) GetAllReactionCounts() map[events.ReactionType]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[events.ReactionType]int64)
	for reactionType, counter := range s.ReactionCounts {
		counts[reactionType] = atomic.LoadInt64(counter)
	}
	return counts
}

// Snapshot returns a complete snapshot of the session statistics
type StatsSnapshot struct {
	SessionID           string                       `json:"session_id"`
	ActiveUserCount     int                          `json:"active_user_count"`
	PeakConcurrentUsers int                          `json:"peak_concurrent_users"`
	TotalReactions      int64                        `json:"total_reactions"`
	ReactionCounts      map[events.ReactionType]int64 `json:"reaction_counts"`
	StartTime           time.Time                    `json:"start_time"`
	LastActivity        time.Time                    `json:"last_activity"`
	Duration            float64                      `json:"duration_seconds"`
}

// GetSnapshot returns a snapshot of the current statistics
func (s *SessionStats) GetSnapshot() StatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return StatsSnapshot{
		SessionID:           s.SessionID,
		ActiveUserCount:     len(s.ActiveUsers),
		PeakConcurrentUsers: s.PeakConcurrentUsers,
		TotalReactions:      atomic.LoadInt64(s.TotalReactions),
		ReactionCounts:      s.GetAllReactionCounts(),
		StartTime:           s.StartTime,
		LastActivity:        s.LastActivity,
		Duration:            time.Since(s.StartTime).Seconds(),
	}
}
