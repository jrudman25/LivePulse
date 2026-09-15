package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// postWithAuth sends an authenticated POST; session create/join now require
// a verified Clerk bearer token (LIVEPULSE_WS_TOKEN or CLERK_JWT).
func postWithAuth(url, token string, body []byte) (*http.Response, error) {
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func main() {
	baseURL := "http://localhost:8080"
	token := os.Getenv("LIVEPULSE_WS_TOKEN")
	if token == "" {
		token = os.Getenv("CLERK_JWT")
	}
	if token == "" {
		log.Println("WARNING: no Clerk token set (LIVEPULSE_WS_TOKEN or CLERK_JWT); session endpoints require auth")
	}

	// 1. Create a session
	fmt.Println("Creating session...")
	sessionReq := map[string]interface{}{
		"name":       "Load Test Event",
		"milestones": []int{10, 50, 100, 500, 1000},
	}

	reqBody, _ := json.Marshal(sessionReq)
	resp, err := postWithAuth(baseURL+"/api/sessions", token, reqBody)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Session creation failed with status %s", resp.Status)
	}

	var sessionResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sessionResp)
	sessionID := sessionResp["session_id"].(string)
	fmt.Printf("Session created: %s\n\n", sessionID)

	// 2. Join the session (identity comes from the verified token subject)
	fmt.Println("Joining session...")
	url := fmt.Sprintf("%s/api/sessions/join?session_id=%s", baseURL, sessionID)
	resp, err = postWithAuth(url, token, nil)
	if err != nil {
		log.Printf("Failed to join session: %v", err)
	} else {
		resp.Body.Close()
		fmt.Printf("  join status: %s\n", resp.Status)
	}
	fmt.Println()

	// 3. Get initial stats
	fmt.Println("📊 Getting initial stats...")
	statsURL := fmt.Sprintf("%s/api/sessions/stats?session_id=%s", baseURL, sessionID)
	resp, err = http.Get(statsURL)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		var stats map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&stats)
		resp.Body.Close()

		statsJSON, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Printf("%s\n\n", statsJSON)
	}

	// 4. Get milestones
	fmt.Println("Getting milestones...")
	milestonesURL := fmt.Sprintf("%s/api/sessions/milestones?session_id=%s", baseURL, sessionID)
	resp, err = http.Get(milestonesURL)
	if err != nil {
		log.Printf("Failed to get milestones: %v", err)
	} else {
		var milestones map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&milestones)
		resp.Body.Close()

		milestonesJSON, _ := json.MarshalIndent(milestones, "", "  ")
		fmt.Printf("%s\n\n", milestonesJSON)
	}

	fmt.Println("Test completed!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Connect via WebSocket to send reactions:")
	fmt.Printf("   ws://localhost:8080/ws?session_id=%s&user_id=user1\n", sessionID)
	fmt.Println("2. Send reaction messages:")
	fmt.Println(`   {"type":"reaction","reaction_type":"like"}`)
	fmt.Println("3. Watch for milestone achievements in the server logs!")
}
