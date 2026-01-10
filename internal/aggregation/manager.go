package aggregation

import (
	"sync"

	"github.com/jrudman25/livepulse/internal/events"
)

// Manager manages statistics for all active sessions
type Manager struct {
	sessions map[string]*SessionStats
	mu       sync.RWMutex
}

// NewManager creates a new aggregation manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*SessionStats),
	}
}

// GetOrCreateSession retrieves or creates session statistics
func (m *Manager) GetOrCreateSession(sessionID string) *SessionStats {
	m.mu.RLock()
	stats, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		return stats
	}

	// Create new session stats
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if stats, exists := m.sessions[sessionID]; exists {
		return stats
	}

	stats = NewSessionStats(sessionID)
	m.sessions[sessionID] = stats
	return stats
}

// GetSession retrieves session statistics if it exists
func (m *Manager) GetSession(sessionID string) (*SessionStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, exists := m.sessions[sessionID]
	return stats, exists
}

// ProcessEvent processes an event and updates statistics
func (m *Manager) ProcessEvent(event *events.Event) {
	stats := m.GetOrCreateSession(event.SessionID)

	switch event.Type {
	case events.EventTypeJoinSession:
		stats.AddUser(event.UserID)
	case events.EventTypeLeaveSession:
		stats.RemoveUser(event.UserID)
	case events.EventTypeReaction:
		if reactionType, ok := event.GetReactionType(); ok {
			stats.IncrementReaction(reactionType)
		}
	}
}

// GetAllSessions returns a snapshot of all session statistics
func (m *Manager) GetAllSessions() map[string]StatsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make(map[string]StatsSnapshot)
	for sessionID, stats := range m.sessions {
		snapshots[sessionID] = stats.GetSnapshot()
	}
	return snapshots
}

// RemoveSession removes a session from tracking
func (m *Manager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// GetSessionCount returns the number of active sessions
func (m *Manager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
