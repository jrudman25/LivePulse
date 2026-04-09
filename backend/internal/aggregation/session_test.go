package aggregation

import (
	"sync"
	"testing"

	"github.com/jrudman25/livepulse/internal/events"
)

func TestSessionStats_ActiveUserConcurrency(t *testing.T) {
	// Initialize a new session tracker
	stats := NewSessionStats("test-session-123")

	// 1. First connection
	count := stats.AddUser("userA")
	if count != 1 {
		t.Errorf("Expected 1 active user, got %d", count)
	}

	// 2. Strict Mode Re-Mount (same user connects a second socket)
	count = stats.AddUser("userA")
	if count != 1 {
		t.Errorf("Expected 1 mathematically distinct active user, got %d", count)
	}
	if stats.ActiveUsers["userA"] != 2 {
		t.Errorf("Expected 2 mathematical overlapping socket connections for userA, got %d", stats.ActiveUsers["userA"])
	}

	// 3. First socket cleanly disconnects
	count = stats.RemoveUser("userA")
	if count != 1 {
		t.Errorf("Expected 1 active user to legally remain connected, got %d", count)
	}
	if stats.ActiveUsers["userA"] != 1 {
		t.Errorf("Expected 1 remaining active socket for userA, got %d", stats.ActiveUsers["userA"])
	}

	// 4. Second socket fully disconnects
	count = stats.RemoveUser("userA")
	if count != 0 {
		t.Errorf("Expected 0 active users remaining, got %d", count)
	}
	if _, exists := stats.ActiveUsers["userA"]; exists {
		t.Errorf("Expected userA to be fully scrubbed from memory array, but key still exists")
	}
}

func TestSessionStats_ConcurrencyStress(t *testing.T) {
	stats := NewSessionStats("test-session-concurrent")
	var wg sync.WaitGroup
	workers := 100

	// Blast 100 concurrent logins
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats.AddUser("userA")
		}()
	}
	wg.Wait()

	if stats.ActiveUsers["userA"] != workers {
		t.Errorf("Expected %d socket states, got %d", workers, stats.ActiveUsers["userA"])
	}
	if stats.GetActiveUserCount() != 1 {
		t.Errorf("Expected 1 distinctly active user, got %d", stats.GetActiveUserCount())
	}

	// Blast 100 concurrent logouts
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats.RemoveUser("userA")
		}()
	}
	wg.Wait()

	if stats.GetActiveUserCount() != 0 {
		t.Errorf("Expected 0 active users after complete teardown, got %d", stats.GetActiveUserCount())
	}
	if _, exists := stats.ActiveUsers["userA"]; exists {
		t.Errorf("Expected map entry to be obliterated gracefully, but it hung in memory")
	}
}

func TestSessionStats_Reactions(t *testing.T) {
	stats := NewSessionStats("test-session-reactions")

	// Trigger custom reactions parallel
	stats.IncrementReaction(events.ReactionFire)
	stats.IncrementReaction(events.ReactionFire)
	stats.IncrementReaction(events.ReactionLove)

	if count := stats.GetReactionCount(events.ReactionFire); count != 2 {
		t.Errorf("Expected 2 Fire reactions, got %d", count)
	}

	if total := stats.GetTotalReactions(); total != 3 {
		t.Errorf("Expected 3 total reactions legally collected, got %d", total)
	}
}
