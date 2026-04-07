package events

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestChatEventCreationAndExtraction(t *testing.T) {
	sessionID := "ticketmaster_123"
	userID := "clerk_456"
	text := "Looking for tickets to the show!"
	authorName := "Jordan"

	// 1. Creation
	event := ChatEvent(sessionID, userID, text, authorName)

	assert.NotNil(t, event, "Event object should securely generate")
	assert.NotEmpty(t, event.ID, "Event UUID should not be blank")
	assert.Equal(t, EventTypeChat, event.Type)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, userID, event.UserID)

	// 2. Data Extraction mapping checks
	extText, extAuthor, ok := event.GetChatText()
	
	assert.True(t, ok, "Extraction should successfully locate payload parameters")
	assert.Equal(t, text, extText)
	assert.Equal(t, authorName, extAuthor)
}

func TestGetChatText_RejectsNonChatEvents(t *testing.T) {
	// Create a structural event simulating internal backend statistics instead of chat
	event := NewEvent(EventTypeReaction, "123", "SYSTEM", map[string]interface{}{
		"active_connections": 50,
	})

	_, _, ok := event.GetChatText()
	assert.False(t, ok, "GetChatText should immediately abort if EventType does not explicitly equal EventTypeChat")
}
