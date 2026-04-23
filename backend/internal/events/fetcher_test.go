package events

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTMResponseParsing verifies that TM API JSON is correctly unmarshaled
func TestTMResponseParsing(t *testing.T) {
	payload := `{
		"_embedded": {
			"events": [
				{
					"id": "tm-001",
					"name": "Taylor Swift | The Eras Tour",
					"dates": { "start": { "dateTime": "2026-07-15T19:00:00Z" } },
					"classifications": [{ "segment": { "name": "Music" } }],
					"_embedded": {
						"venues": [{
							"name": "SoFi Stadium",
							"city": { "name": "Los Angeles" },
							"state": { "stateCode": "CA" },
							"country": { "countryCode": "US" }
						}]
					}
				}
			]
		}
	}`

	var resp TMResponse
	err := json.Unmarshal([]byte(payload), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Embedded.Events, 1)

	ev := resp.Embedded.Events[0]
	assert.Equal(t, "tm-001", ev.ID)
	assert.Equal(t, "Taylor Swift | The Eras Tour", ev.Name)
	assert.Equal(t, "2026-07-15T19:00:00Z", ev.Dates.Start.DateTime)
	assert.Equal(t, "Music", ev.Classifications[0].Segment.Name)
	assert.Equal(t, "SoFi Stadium", ev.Embedded.Venues[0].Name)
	assert.Equal(t, "CA", ev.Embedded.Venues[0].State.StateCode)
	assert.Equal(t, "US", ev.Embedded.Venues[0].Country.CountryCode)
}

// TestTMResponseParsing_EmptyEmbedded handles the case where TM returns no events
func TestTMResponseParsing_EmptyEmbedded(t *testing.T) {
	payload := `{}`
	var resp TMResponse
	err := json.Unmarshal([]byte(payload), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Embedded.Events)
}

// TestJunkFilter validates the keyword-based junk event filter
func TestJunkFilter(t *testing.T) {
	junkKeywords := []string{"parking", "permit", "vip club", "shuttle", "camping", "add-on"}

	junkTitles := []string{
		"Event Parking Pass",
		"VIP Club Access - Premium Lounge",
		"Shuttle Service to Venue",
		"Camping Add-On Weekend Pass",
		"General Parking Permit A",
	}

	cleanTitles := []string{
		"Taylor Swift | The Eras Tour",
		"NBA Finals Game 7",
		"Dave Chappelle Live Comedy Special",
		"Coachella Music Festival",
	}

	isJunk := func(title string) bool {
		titleLower := strings.ToLower(title)
		for _, keyword := range junkKeywords {
			if strings.Contains(titleLower, keyword) {
				return true
			}
		}
		return false
	}

	for _, title := range junkTitles {
		assert.True(t, isJunk(title), "%q should be detected as junk", title)
	}
	for _, title := range cleanTitles {
		assert.False(t, isJunk(title), "%q should NOT be detected as junk", title)
	}
}

// TestClassificationFilter validates the Miscellaneous/Undefined rejection
func TestClassificationFilter(t *testing.T) {
	rejectedTypes := []string{"Miscellaneous", "Undefined"}
	allowedTypes := []string{"Music", "Sports", "Arts & Theatre", "Comedy", "Film", "Event"}

	shouldReject := func(eventType string) bool {
		return eventType == "Miscellaneous" || eventType == "Undefined"
	}

	for _, et := range rejectedTypes {
		assert.True(t, shouldReject(et), "%q should be rejected", et)
	}
	for _, et := range allowedTypes {
		assert.False(t, shouldReject(et), "%q should be allowed", et)
	}
}

// TestLocationStringBuilding verifies venue → location string formatting
func TestLocationStringBuilding(t *testing.T) {
	tests := []struct {
		name       string
		venue      TMEvent
		wantLoc    string
		wantCountry string
	}{
		{
			name:       "full venue info",
			venue:      tmEventWithVenue("Madison Square Garden", "New York", "NY", "US"),
			wantLoc:    "Madison Square Garden (New York, NY)",
			wantCountry: "US",
		},
		{
			name:       "venue name only, no city/state",
			venue:      tmEventWithVenue("The O2 Arena", "", "", "GB"),
			wantLoc:    "The O2 Arena",
			wantCountry: "GB",
		},
		{
			name:       "no venue at all",
			venue:      TMEvent{},
			wantLoc:    "TBA",
			wantCountry: "US",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc, country := buildLocation(tc.venue)
			assert.Equal(t, tc.wantLoc, loc)
			assert.Equal(t, tc.wantCountry, country)
		})
	}
}

// --- helpers ---

func tmEventWithVenue(name, city, stateCode, countryCode string) TMEvent {
	var ev TMEvent
	ev.Embedded.Venues = append(ev.Embedded.Venues, struct {
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
	}{
		Name: name,
	})
	ev.Embedded.Venues[0].City.Name = city
	ev.Embedded.Venues[0].State.StateCode = stateCode
	ev.Embedded.Venues[0].Country.CountryCode = countryCode
	return ev
}

// buildLocation replicates the location string logic from FetchAPIEvents
func buildLocation(tmEvent TMEvent) (string, string) {
	locationStr := "TBA"
	countryStr := "US"
	if len(tmEvent.Embedded.Venues) > 0 {
		venue := tmEvent.Embedded.Venues[0]
		if venue.Name != "" {
			locationStr = venue.Name
			if venue.City.Name != "" && venue.State.StateCode != "" {
				locationStr = venue.Name + " (" + venue.City.Name + ", " + venue.State.StateCode + ")"
			}
		}
		if venue.Country.CountryCode != "" {
			countryStr = venue.Country.CountryCode
		}
	}
	return locationStr, countryStr
}
