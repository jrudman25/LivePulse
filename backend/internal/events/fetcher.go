package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	// Schedule to run every 6 hours to keep events fresh
	_, err := f.cron.AddFunc("0 */6 * * *", f.FetchAPIEvents)
	if err != nil {
		log.Printf("Error scheduling event fetcher: %v", err)
		return
	}
	f.cron.Start()
	log.Println("Event API fetcher started via cron: scheduled every 6 hours")

	// Run an initial fetch on startup so events are available immediately
	go f.FetchAPIEvents()
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
	Embedded struct {
		Venues []struct {
			Name string `json:"name"`
			City struct {
				Name string `json:"name"`
			} `json:"city"`
			State struct {
				StateCode string `json:"stateCode"`
			} `json:"state"`
			Country struct {
				CountryCode string `json:"countryCode"`
			} `json:"country"`
		} `json:"venues"`
	} `json:"_embedded"`
}

// FetchAPIEvents hits the Ticketmaster API and populates the DB
func (f *APIFetcher) FetchAPIEvents() {
	if f.apiKey == "" || f.apiKey == "your_ticketmaster_api_key" {
		log.Println("Skipping TM ingestion: EXTERNAL_API_KEY is missing or set to default")
		return
	}

	log.Println("Fetching new events from Ticketmaster API...")

	now := time.Now().UTC()
	nowStr := now.Format("2006-01-02T15:04:05Z")
	endStr := now.Add(24 * time.Hour).Format("2006-01-02T15:04:05Z")

	// Lock fetches perfectly onto verified high-quality event categories natively formatted for URL consumption
	classificationParams := "classificationName=Music,Sports,Arts%20%26%20Theatre,Comedy,Film"

	for page := 0; page < 2; page++ {
		exitLoop := func() bool {
			urlQuery := fmt.Sprintf("https://app.ticketmaster.com/discovery/v2/events.json?apikey=%s&size=200&page=%d&sort=relevance,desc&startDateTime=%s&endDateTime=%s&%s", f.apiKey, page, nowStr, endStr, classificationParams)
			resp, err := http.Get(urlQuery)
			if err != nil {
				log.Printf("Error requesting TM API on page %d: %v", page, err)
				return true
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("TM API returned status code %d on page %d (likely exhausted Timeline events), halting pagination.", resp.StatusCode, page)
				return true
			}

			var tmResp TMResponse
			if err := json.NewDecoder(resp.Body).Decode(&tmResp); err != nil {
				log.Printf("Error decoding TM API JSON: %v", err)
				return true
			}

			for _, tmEvent := range tmResp.Embedded.Events {
				startTime, err := time.Parse(time.RFC3339, tmEvent.Dates.Start.DateTime)
				if err != nil {
					// Skip generic events that lack a rigid start date entirely
					log.Printf("Skipping event with invalid date: %s", tmEvent.Name)
					continue
				}

				eventType := "Event"
				if len(tmEvent.Classifications) > 0 {
					eventType = tmEvent.Classifications[0].Segment.Name
				}

				// Aggressive fallback safety filter rejecting generic "Miscellaneous"
				importType := eventType
				if importType == "Miscellaneous" || importType == "Undefined" {
					continue
				}

				locationStr := "TBA"
				countryStr := "US"
				if len(tmEvent.Embedded.Venues) > 0 {
					venue := tmEvent.Embedded.Venues[0]
					if venue.Name != "" {
						locationStr = venue.Name
						if venue.City.Name != "" && venue.State.StateCode != "" {
							locationStr = fmt.Sprintf("%s (%s, %s)", venue.Name, venue.City.Name, venue.State.StateCode)
						}
					}
					if venue.Country.CountryCode != "" {
						countryStr = venue.Country.CountryCode
					}
				}

				// Database logic will purge legacy data. Future fetched data is shielded organically above by classification limits.
				titleLower := strings.ToLower(tmEvent.Name)
				isJunk := false
				for _, keyword := range []string{"parking", "permit", "vip club", "shuttle", "camping", "add-on"} {
					if strings.Contains(titleLower, keyword) {
						isJunk = true
						break
					}
				}
				if isJunk {
					log.Printf("Skipping metadata junk pass: %s", tmEvent.Name)
					continue
				}

				e := storage.Event{
					ID:            tmEvent.ID,
					Type:          eventType,
					Title:         tmEvent.Name,
					Location:      locationStr,
					Country:       countryStr,
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

			return false
		}()

		if exitLoop {
			break
		}
	}

	// Trigger PostgreSQL garbage collector to natively erase dead history mapping explicitly to -1 Hour constraints
	if err := f.db.DeleteExpiredEvents(context.Background()); err != nil {
		log.Printf("Error deleting expired events from DB: %v", err)
	} else {
		log.Println("Pristinely swept database of legacy expired events.")
	}
}

// FetchSearchKeyword provides infinite On-Demand database insertion by aggressively searching Ticketmaster natively
func (f *APIFetcher) FetchSearchKeyword(keyword string) {
	if f.apiKey == "" || f.apiKey == "your_ticketmaster_api_key" {
		return
	}

	nowStr := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	safeKeyword := url.QueryEscape(keyword)
	urlQuery := fmt.Sprintf("https://app.ticketmaster.com/discovery/v2/events.json?apikey=%s&size=50&sort=date,asc&startDateTime=%s&keyword=%s", f.apiKey, nowStr, safeKeyword)

	resp, err := http.Get(urlQuery)
	if err != nil {
		log.Printf("Error requesting TM search API: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var tmResp TMResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmResp); err != nil {
		return
	}

	for _, tmEvent := range tmResp.Embedded.Events {
		startTime, err := time.Parse(time.RFC3339, tmEvent.Dates.Start.DateTime)
		if err != nil {
			continue
		}

		eventType := "Event"
		if len(tmEvent.Classifications) > 0 {
			eventType = tmEvent.Classifications[0].Segment.Name
		}

		locationStr := "TBA"
		countryStr := "US"
		if len(tmEvent.Embedded.Venues) > 0 {
			venue := tmEvent.Embedded.Venues[0]
			if venue.Name != "" {
				locationStr = venue.Name
				if venue.City.Name != "" && venue.State.StateCode != "" {
					locationStr = fmt.Sprintf("%s (%s, %s)", venue.Name, venue.City.Name, venue.State.StateCode)
				}
			}
			if venue.Country.CountryCode != "" {
				countryStr = venue.Country.CountryCode
			}
		}

		e := storage.Event{
			ID:            tmEvent.ID,
			Type:          eventType,
			Title:         tmEvent.Name,
			Location:      locationStr,
			Country:       countryStr,
			StartTime:     startTime,
			EndTime:       startTime.Add(3 * time.Hour),
			ExternalAPIID: tmEvent.ID,
			CreatedAt:     time.Now(),
		}

		// Insert silently to maintain user velocity!
		f.db.InsertEvent(context.Background(), e)
	}
}
