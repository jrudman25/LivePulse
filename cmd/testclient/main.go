package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8080"

	// 1. Create a session
	fmt.Println("Creating session...")
	sessionReq := map[string]interface{}{
		"name":       "Load Test Event",
		"milestones": []int{10, 50, 100, 500, 1000},
	}

	reqBody, _ := json.Marshal(sessionReq)
	resp, err := http.Post(baseURL+"/api/sessions", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer resp.Body.Close()

	var sessionResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sessionResp)
	sessionID := sessionResp["session_id"].(string)
	fmt.Printf("Session created: %s\n\n", sessionID)

	// 2. Simulate multiple users joining
	fmt.Println("Simulating users joining...")
	for i := 1; i <= 10; i++ {
		userID := fmt.Sprintf("user%d", i)
		url := fmt.Sprintf("%s/api/sessions/join?session_id=%s&user_id=%s", baseURL, sessionID, userID)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			log.Printf("Failed to join user %s: %v", userID, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("  ✓ %s joined\n", userID)
		time.Sleep(100 * time.Millisecond)
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
