package milestones

import (
	"log"
	"sync"
	"time"

	"github.com/jrudman25/livepulse/internal/aggregation"
)

// NotificationHandler is called when a milestone is achieved
type NotificationHandler func(*MilestoneAchievement)

// Tracker tracks milestones for sessions
type Tracker struct {
	milestones map[string][]*Milestone // sessionID -> milestones
	mu         sync.RWMutex
	notifyFunc NotificationHandler
}

// NewTracker creates a new milestone tracker
func NewTracker(notifyFunc NotificationHandler) *Tracker {
	return &Tracker{
		milestones: make(map[string][]*Milestone),
		notifyFunc: notifyFunc,
	}
}

// InitializeSession sets up milestones for a session
func (t *Tracker) InitializeSession(sessionID string, thresholds []int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var milestones []*Milestone

	// Create reaction milestones
	for _, threshold := range thresholds {
		milestone := NewMilestone(sessionID, MilestoneTypeTotalReactions, int64(threshold))
		milestones = append(milestones, milestone)
	}

	t.milestones[sessionID] = milestones
	log.Printf("Initialized %d milestones for session %s", len(milestones), sessionID)
}

// CheckMilestones checks if any milestones were achieved based on current stats
func (t *Tracker) CheckMilestones(sessionID string, stats *aggregation.SessionStats) {
	totalReactions := stats.GetTotalReactions()
	activeUsers := int64(stats.GetActiveUserCount())

	// Hold the write lock for the whole check/update pass so concurrent
	// workers cannot race on Progress/Achieved, then notify after unlock.
	t.mu.Lock()
	sessionMilestones, exists := t.milestones[sessionID]
	if !exists {
		t.mu.Unlock()
		return
	}

	var achievements []*MilestoneAchievement
	for _, milestone := range sessionMilestones {
		if milestone.Achieved {
			continue // Already achieved
		}

		var currentValue int64
		switch milestone.Type {
		case MilestoneTypeTotalReactions:
			currentValue = totalReactions
		case MilestoneTypeConcurrentUsers:
			currentValue = activeUsers
		case MilestoneTypeSessionDuration:
			currentValue = int64(time.Since(stats.StartTime).Minutes())
		}

		// Update progress and check if just achieved
		if milestone.UpdateProgress(currentValue) {
			achievements = append(achievements, &MilestoneAchievement{
				Milestone:    *milestone, // snapshot while locked so marshaling never races
				SessionID:    sessionID,
				AchievedAt:   time.Now().UTC(),
				CurrentValue: currentValue,
			})

			log.Printf("Milestone achieved! Session: %s, Type: %s, Threshold: %d, Current: %d",
				sessionID, milestone.Type, milestone.Threshold, currentValue)
		}
	}
	t.mu.Unlock()

	for _, achievement := range achievements {
		if t.notifyFunc != nil {
			go t.notifyFunc(achievement)
		}
	}
}

// GetSessionMilestones returns copies of all milestones for a session
func (t *Tracker) GetSessionMilestones(sessionID string) []Milestone {
	t.mu.RLock()
	defer t.mu.RUnlock()

	milestones, exists := t.milestones[sessionID]
	if !exists {
		return nil
	}

	result := make([]Milestone, len(milestones))
	for i, m := range milestones {
		result[i] = *m
	}
	return result
}

// GetAchievedMilestones returns copies of the achieved milestones for a session
func (t *Tracker) GetAchievedMilestones(sessionID string) []Milestone {
	t.mu.RLock()
	defer t.mu.RUnlock()

	milestones, exists := t.milestones[sessionID]
	if !exists {
		return nil
	}

	var achieved []Milestone
	for _, m := range milestones {
		if m.Achieved {
			achieved = append(achieved, *m)
		}
	}
	return achieved
}

// RemoveSession removes milestone tracking for a session
func (t *Tracker) RemoveSession(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.milestones, sessionID)
}

// AddCustomMilestone adds a custom milestone to a session
func (t *Tracker) AddCustomMilestone(sessionID string, milestoneType MilestoneType, threshold int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	milestone := NewMilestone(sessionID, milestoneType, threshold)
	t.milestones[sessionID] = append(t.milestones[sessionID], milestone)
}
