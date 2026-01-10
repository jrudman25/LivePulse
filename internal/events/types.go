package events

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeJoinSession  EventType = "join_session"
	EventTypeLeaveSession EventType = "leave_session"
	EventTypeReaction     EventType = "reaction"
)

// ReactionType represents different types of reactions
type ReactionType string

const (
	ReactionLike     ReactionType = "like"
	ReactionLove     ReactionType = "love"
	ReactionCheer    ReactionType = "cheer"
	ReactionApplause ReactionType = "applause"
	ReactionFire     ReactionType = "fire"
	ReactionHeart    ReactionType = "heart"
)

// Event represents a user action in a session
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewEvent creates a new event with a generated ID and timestamp
func NewEvent(eventType EventType, sessionID, userID string, payload map[string]interface{}) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		SessionID: sessionID,
		UserID:    userID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// ReactionEvent creates a reaction event
func ReactionEvent(sessionID, userID string, reactionType ReactionType) *Event {
	return NewEvent(EventTypeReaction, sessionID, userID, map[string]interface{}{
		"reaction_type": reactionType,
	})
}

// JoinSessionEvent creates a join session event
func JoinSessionEvent(sessionID, userID string) *Event {
	return NewEvent(EventTypeJoinSession, sessionID, userID, nil)
}

// LeaveSessionEvent creates a leave session event
func LeaveSessionEvent(sessionID, userID string) *Event {
	return NewEvent(EventTypeLeaveSession, sessionID, userID, nil)
}

// GetReactionType extracts the reaction type from the event payload
func (e *Event) GetReactionType() (ReactionType, bool) {
	if e.Type != EventTypeReaction {
		return "", false
	}
	if rt, ok := e.Payload["reaction_type"].(string); ok {
		return ReactionType(rt), true
	}
	return "", false
}
