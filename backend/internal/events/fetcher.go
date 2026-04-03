package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// TMResponse models the Ticketmaster JSON payload
type TMResponse struct {
	Embedded struct {
		Events []TMEvent `json:"events"`
	} `json:"_embedded"`
}

type TMEvent struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Dates struct {
		Start struct {
			DateTime string `json:"dateTime"`
		} `json:"start"`
	} `json:"dates"`
	Classifications []struct {
		Segment struct {
			Name string `json:"name"`
		} `json:"segment"`
	} `json:"classifications"`
}

// FetchAPIEvents hits the Ticketmaster API and populates the DB
func (f *APIFetcher) FetchAPIEvents() {
	if f.apiKey == "" || f.apiKey == "your_ticketmaster_api_key" {
		log.Println("Skipping TM ingestion: EXTERNAL_API_KEY is missing or set to default")
		return
	}

	log.Println("Fetching new events from Ticketmaster API...")

	nowStr := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	url := fmt.Sprintf("https://app.ticketmaster.com/discovery/v2/events.json?apikey=%s&size=30&sort=date,asc&startDateTime=%s", f.apiKey, nowStr)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error requesting TM API: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("TM API returned status code: %d", resp.StatusCode)
		return
	}

	var tmResp TMResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmResp); err != nil {
		log.Printf("Error decoding TM API JSON: %v", err)
		return
	}

	for _, tmEvent := range tmResp.Embedded.Events {
		startTime, err := time.Parse(time.RFC3339, tmEvent.Dates.Start.DateTime)
		if err != nil {
			// Fallback to exactly tomorrow if TM date is missing or malformed
			startTime = time.Now().Add(24 * time.Hour) 
		}

		eventType := "Event"
		if len(tmEvent.Classifications) > 0 {
			eventType = tmEvent.Classifications[0].Segment.Name
		}

		e := storage.Event{
			ID:            uuid.New().String(),
			Type:          eventType,
			Title:         tmEvent.Name,
			StartTime:     startTime,
			EndTime:       startTime.Add(3 * time.Hour), // TM strict end-times are often missing, 3 hours is a safe heuristic
			ExternalAPIID: tmEvent.ID,
			CreatedAt:     time.Now(),
		}

		if err := f.db.InsertEvent(context.Background(), e); err != nil {
			log.Printf("Failed to insert event %s: %v", e.ExternalAPIID, err)
		} else {
			log.Printf("Successfully ingested event: %s", e.Title)
		}
	}
}
