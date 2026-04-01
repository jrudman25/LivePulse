package events

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jrudman25/livepulse/internal/storage"
	"github.com/robfig/cron/v3"
)

// APIFetcher handles background event ingestion
type APIFetcher struct {
	db     *storage.PostgresClient
	cron   *cron.Cron
	apiKey string
}

// NewAPIFetcher initializes the background fetch scheduler
func NewAPIFetcher(db *storage.PostgresClient, apiKey string) *APIFetcher {
	c := cron.New()
	return &APIFetcher{
		db:     db,
		cron:   c,
		apiKey: apiKey,
	}
}

// Start begins the daily cron jobs to fetch events
func (f *APIFetcher) Start() {
	// Schedule to run every day at 2:00 AM
	_, err := f.cron.AddFunc("0 2 * * *", f.FetchAPIEvents)
	if err != nil {
		log.Printf("Error scheduling event fetcher: %v", err)
		return
	}
	f.cron.Start()
	log.Println("Event API fetcher started via cron: scheduled at 02:00 AM daily")
}

// Stop halts the cron scheduler
func (f *APIFetcher) Stop() {
	f.cron.Stop()
}

// FetchAPIEvents hits a generic API and populates the DB
func (f *APIFetcher) FetchAPIEvents() {
	log.Println("Fetching new events from external API...")

	// In a complete implementation, this would use f.apiKey to hit Ticketmaster, SeatGeek, etc.
	// We simulate a parsed response payload for demonstration purposes.
	mockEvents := []storage.Event{
		{
			ID:            uuid.New().String(),
			Type:          "concert",
			Title:         "Sample Concert Main Event",
			StartTime:     time.Now().Add(24 * time.Hour),
			EndTime:       time.Now().Add(27 * time.Hour),
			ExternalAPIID: "tm_dummy_001",
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New().String(),
			Type:          "sports",
			Title:         "Championship Game",
			StartTime:     time.Now().Add(48 * time.Hour),
			EndTime:       time.Now().Add(52 * time.Hour),
			ExternalAPIID: "tm_dummy_002",
			CreatedAt:     time.Now(),
		},
	}

	for _, e := range mockEvents {
		err := f.db.InsertEvent(context.Background(), e)
		if err != nil {
			log.Printf("Failed to insert event %s: %v", e.ExternalAPIID, err)
		} else {
			log.Printf("Successfully ingested event: %s", e.Title)
		}
	}
}
