package milestones

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jrudman25/livepulse/internal/aggregation"
	"github.com/jrudman25/livepulse/internal/events"
)

// TestTracker_CheckMilestonesConcurrent runs concurrent milestone checks
// from multiple workers against one session. Run with -race: before the
// tracker serialized updates under its write lock, this raced on
// Progress/Achieved and could double-fire achievements.
func TestTracker_CheckMilestonesConcurrent(t *testing.T) {
	var notifications int64
	tracker := NewTracker(func(*MilestoneAchievement) {
		atomic.AddInt64(&notifications, 1)
	})

	sessionID := "test-session-milestones"
	tracker.InitializeSession(sessionID, []int{100})

	stats := aggregation.NewSessionStats(sessionID)

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				stats.IncrementReaction(events.ReactionFire)
				tracker.CheckMilestones(sessionID, stats)
			}
		}()
	}
	wg.Wait()

	// The threshold was crossed exactly once; the milestone must fire once.
	if got := atomic.LoadInt64(&notifications); got != 1 {
		t.Errorf("Expected exactly 1 achievement notification, got %d", got)
	}

	achieved := tracker.GetAchievedMilestones(sessionID)
	if len(achieved) != 1 {
		t.Fatalf("Expected 1 achieved milestone, got %d", len(achieved))
	}
	if !achieved[0].Achieved {
		t.Error("Returned milestone should be marked achieved")
	}
}

// TestTracker_GetSessionMilestonesReturnsCopies verifies callers cannot
// mutate tracker state through returned milestones.
func TestTracker_GetSessionMilestonesReturnsCopies(t *testing.T) {
	tracker := NewTracker(nil)
	tracker.InitializeSession("s", []int{10})

	got := tracker.GetSessionMilestones("s")
	if len(got) != 1 {
		t.Fatalf("Expected 1 milestone, got %d", len(got))
	}
	got[0].Achieved = true
	got[0].Progress = 9999

	again := tracker.GetSessionMilestones("s")
	if again[0].Achieved || again[0].Progress != 0 {
		t.Error("GetSessionMilestones must return value copies, not shared state")
	}
}
