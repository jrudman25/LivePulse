package milestones

import (
	"strconv"
	"time"
)

// MilestoneType represents different types of milestones
type MilestoneType string

const (
	MilestoneTypeTotalReactions  MilestoneType = "total_reactions"
	MilestoneTypeConcurrentUsers MilestoneType = "concurrent_users"
	MilestoneTypeSessionDuration MilestoneType = "session_duration"
)

// Milestone represents a goal that can be achieved
type Milestone struct {
	ID          string        `json:"id"`
	SessionID   string        `json:"session_id"`
	Type        MilestoneType `json:"type"`
	Threshold   int64         `json:"threshold"`
	Progress    int64         `json:"progress"`
	Achieved    bool          `json:"achieved"`
	AchievedAt  *time.Time    `json:"achieved_at,omitempty"`
	Description string        `json:"description"`
}

// MilestoneAchievement represents a milestone that was just achieved
type MilestoneAchievement struct {
	Milestone    *Milestone `json:"milestone"`
	SessionID    string     `json:"session_id"`
	AchievedAt   time.Time  `json:"achieved_at"`
	CurrentValue int64      `json:"current_value"`
}

// NewMilestone creates a new milestone
func NewMilestone(sessionID string, milestoneType MilestoneType, threshold int64) *Milestone {
	return &Milestone{
		ID:          generateMilestoneID(sessionID, milestoneType, threshold),
		SessionID:   sessionID,
		Type:        milestoneType,
		Threshold:   threshold,
		Progress:    0,
		Achieved:    false,
		Description: generateDescription(milestoneType, threshold),
	}
}

// generateMilestoneID creates a unique ID for a milestone
func generateMilestoneID(sessionID string, milestoneType MilestoneType, threshold int64) string {
	return sessionID + "_" + string(milestoneType) + "_" + strconv.FormatInt(threshold, 10)
}

// generateDescription creates a human-readable description
func generateDescription(milestoneType MilestoneType, threshold int64) string {
	switch milestoneType {
	case MilestoneTypeTotalReactions:
		return formatNumber(threshold) + " total reactions"
	case MilestoneTypeConcurrentUsers:
		return formatNumber(threshold) + " concurrent users"
	case MilestoneTypeSessionDuration:
		return formatNumber(threshold) + " minutes session duration"
	default:
		return "Unknown milestone"
	}
}

// formatNumber formats a number with commas for readability
func formatNumber(n int64) string {
	return strconv.FormatInt(n, 10)
}

// UpdateProgress updates the milestone progress
func (m *Milestone) UpdateProgress(currentValue int64) bool {
	m.Progress = currentValue

	if !m.Achieved && currentValue >= m.Threshold {
		m.Achieved = true
		now := time.Now().UTC()
		m.AchievedAt = &now
		return true // Milestone just achieved
	}

	return false
}

// ProgressPercentage returns the progress as a percentage
func (m *Milestone) ProgressPercentage() float64 {
	if m.Threshold == 0 {
		return 0
	}
	percentage := (float64(m.Progress) / float64(m.Threshold)) * 100
	if percentage > 100 {
		return 100
	}
	return percentage
}
