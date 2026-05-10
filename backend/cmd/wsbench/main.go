package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type config struct {
	wsURL             string
	sessionID         string
	token             string
	origin            string
	clients           int
	messagesPerClient int
	warmupMessages    int
	timeout           time.Duration
	authorName        string
}

type runResult struct {
	authLatencies []time.Duration
	chatLatencies []time.Duration
	failures      int64
	errors        []error
	elapsed       time.Duration
}

type chatEnvelope struct {
	Type    string       `json:"type"`
	Message *chatMessage `json:"message,omitempty"`
	Error   string
}

type chatMessage struct {
	Text string `json:"text"`
}

type rawEnvelope struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message,omitempty"`
}

func main() {
	cfg := parseConfig()

	if cfg.warmupMessages > 0 {
		log.Printf("Running warmup: clients=%d messages_per_client=%d", cfg.clients, cfg.warmupMessages)
		warmupCfg := cfg
		warmupCfg.messagesPerClient = cfg.warmupMessages
		result := runBenchmark(warmupCfg)
		if len(result.errors) > 0 {
			log.Printf("Warmup completed with %d errors", len(result.errors))
		}
	}

	log.Printf("Running benchmark: url=%s session=%s clients=%d messages_per_client=%d", cfg.wsURL, cfg.sessionID, cfg.clients, cfg.messagesPerClient)
	result := runBenchmark(cfg)
	printResult(cfg, result)

	if len(result.errors) > 0 {
		os.Exit(1)
	}
}

func parseConfig() config {
	cfg := config{}
	flag.StringVar(&cfg.wsURL, "url", "ws://localhost:8080", "WebSocket base URL or full /ws URL")
	flag.StringVar(&cfg.sessionID, "session", "wsbench", "session_id query parameter")
	flag.StringVar(&cfg.token, "token", firstNonEmpty(os.Getenv("LIVEPULSE_WS_TOKEN"), os.Getenv("CLERK_JWT")), "Clerk JWT token, defaults to LIVEPULSE_WS_TOKEN or CLERK_JWT")
	flag.StringVar(&cfg.origin, "origin", "", "optional Origin header")
	flag.IntVar(&cfg.clients, "clients", 1, "number of concurrent WebSocket clients")
	flag.IntVar(&cfg.messagesPerClient, "messages", 10, "chat messages sent by each client")
	flag.IntVar(&cfg.warmupMessages, "warmup", 0, "warmup messages per client before measured run")
	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "read/write timeout per operation")
	flag.StringVar(&cfg.authorName, "author", "WS Bench", "author_name sent with chat messages")
	flag.Parse()

	if cfg.token == "" {
		log.Fatal("missing Clerk JWT: set LIVEPULSE_WS_TOKEN or CLERK_JWT, or pass -token")
	}
	if cfg.clients <= 0 {
		log.Fatal("clients must be greater than 0")
	}
	if cfg.messagesPerClient <= 0 {
		log.Fatal("messages must be greater than 0")
	}
	if cfg.timeout <= 0 {
		log.Fatal("timeout must be greater than 0")
	}

	return cfg
}

func runBenchmark(cfg config) runResult {
	started := time.Now()
	endpoint, err := endpointForSession(cfg.wsURL, cfg.sessionID)
	if err != nil {
		return runResult{errors: []error{err}}
	}

	authLatencies := make(chan time.Duration, cfg.clients)
	chatLatencies := make(chan time.Duration, cfg.clients*cfg.messagesPerClient)
	errCh := make(chan error, cfg.clients*cfg.messagesPerClient+cfg.clients)
	runID := fmt.Sprintf("%d", time.Now().UnixNano())

	var failures int64
	var wg sync.WaitGroup
	wg.Add(cfg.clients)

	for clientID := 0; clientID < cfg.clients; clientID++ {
		go func(clientID int) {
			defer wg.Done()
			if err := runClient(cfg, endpoint, runID, clientID, authLatencies, chatLatencies); err != nil {
				atomic.AddInt64(&failures, 1)
				errCh <- err
			}
		}(clientID)
	}

	wg.Wait()
	close(authLatencies)
	close(chatLatencies)
	close(errCh)

	result := runResult{
		authLatencies: drainDurations(authLatencies),
		chatLatencies: drainDurations(chatLatencies),
		failures:      failures,
		elapsed:       time.Since(started),
	}
	for err := range errCh {
		result.errors = append(result.errors, err)
	}
	return result
}

func runClient(cfg config, endpoint string, runID string, clientID int, authLatencies chan<- time.Duration, chatLatencies chan<- time.Duration) error {
	dialer := websocket.Dialer{HandshakeTimeout: cfg.timeout}
	header := http.Header{}
	if cfg.origin != "" {
		header.Set("Origin", cfg.origin)
	}

	authStarted := time.Now()
	conn, _, err := dialer.Dial(endpoint, header)
	if err != nil {
		return fmt.Errorf("client %d dial: %w", clientID, err)
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(cfg.timeout)); err != nil {
		return fmt.Errorf("client %d auth write deadline: %w", clientID, err)
	}
	if err := conn.WriteJSON(map[string]string{"type": "authenticate", "token": cfg.token}); err != nil {
		return fmt.Errorf("client %d auth write: %w", clientID, err)
	}
	if err := waitForAuthenticated(conn, cfg.timeout); err != nil {
		return fmt.Errorf("client %d auth read: %w", clientID, err)
	}
	authLatencies <- time.Since(authStarted)

	for i := 0; i < cfg.messagesPerClient; i++ {
		text := fmt.Sprintf("wsbench:%s:%d:%d", runID, clientID, i)
		payload := map[string]string{
			"type":        "chat",
			"text":        text,
			"author_name": cfg.authorName,
		}

		messageStarted := time.Now()
		if err := conn.SetWriteDeadline(time.Now().Add(cfg.timeout)); err != nil {
			return fmt.Errorf("client %d message %d write deadline: %w", clientID, i, err)
		}
		if err := conn.WriteJSON(payload); err != nil {
			return fmt.Errorf("client %d message %d write: %w", clientID, i, err)
		}
		if err := waitForChatEcho(conn, cfg.timeout, text); err != nil {
			return fmt.Errorf("client %d message %d read: %w", clientID, i, err)
		}
		chatLatencies <- time.Since(messageStarted)
	}

	return nil
}

func waitForAuthenticated(conn *websocket.Conn, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return err
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		for _, frame := range splitFrames(message) {
			payload, err := parseEnvelope(frame)
			if err != nil {
				continue
			}
			switch payload.Type {
			case "authenticated":
				return nil
			case "error":
				return errors.New(payload.Error)
			}
		}
	}
	return errors.New("authentication response timed out")
}

func waitForChatEcho(conn *websocket.Conn, timeout time.Duration, expectedText string) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return err
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		for _, frame := range splitFrames(message) {
			payload, err := parseEnvelope(frame)
			if err != nil {
				continue
			}
			if payload.Type == "error" {
				return errors.New(payload.Error)
			}
			if payload.Type == "chat" && payload.Message != nil && payload.Message.Text == expectedText {
				return nil
			}
		}
	}
	return fmt.Errorf("chat response timed out for %q", expectedText)
}

func parseEnvelope(frame []byte) (chatEnvelope, error) {
	var raw rawEnvelope
	if err := json.Unmarshal(frame, &raw); err != nil {
		return chatEnvelope{}, err
	}

	payload := chatEnvelope{Type: raw.Type}
	if raw.Type == "error" {
		var message string
		if err := json.Unmarshal(raw.Message, &message); err == nil {
			payload.Error = message
		}
		return payload, nil
	}

	if raw.Type == "chat" {
		var message chatMessage
		if err := json.Unmarshal(raw.Message, &message); err == nil {
			payload.Message = &message
		}
	}

	return payload, nil
}

func splitFrames(message []byte) [][]byte {
	parts := bytes.Split(message, []byte("\n"))
	frames := make([][]byte, 0, len(parts))
	for _, part := range parts {
		trimmed := bytes.TrimSpace(part)
		if len(trimmed) > 0 {
			frames = append(frames, trimmed)
		}
	}
	return frames
}

func endpointForSession(rawURL string, sessionID string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
	}

	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/ws"
	}

	query := parsed.Query()
	query.Set("session_id", sessionID)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func drainDurations(ch <-chan time.Duration) []time.Duration {
	values := make([]time.Duration, 0)
	for value := range ch {
		values = append(values, value)
	}
	return values
}

func printResult(cfg config, result runResult) {
	totalMessages := cfg.clients * cfg.messagesPerClient
	successfulMessages := len(result.chatLatencies)
	throughput := 0.0
	if result.elapsed > 0 {
		throughput = float64(successfulMessages) / result.elapsed.Seconds()
	}

	fmt.Println()
	fmt.Println("LivePulse WebSocket Benchmark Results")
	fmt.Println(strings.Repeat("=", 39))
	fmt.Printf("Clients:              %d\n", cfg.clients)
	fmt.Printf("Messages per client:  %d\n", cfg.messagesPerClient)
	fmt.Printf("Attempted messages:   %d\n", totalMessages)
	fmt.Printf("Successful messages:  %d\n", successfulMessages)
	fmt.Printf("Client failures:      %d\n", result.failures)
	fmt.Printf("Elapsed:              %s\n", result.elapsed.Round(time.Millisecond))
	fmt.Printf("Throughput:           %.2f messages/sec\n", throughput)
	fmt.Println()
	printStats("Auth handshake latency", result.authLatencies)
	fmt.Println()
	printStats("Chat round-trip latency", result.chatLatencies)

	if len(result.errors) > 0 {
		fmt.Println()
		fmt.Println("Errors")
		fmt.Println(strings.Repeat("-", 6))
		maxErrors := len(result.errors)
		if maxErrors > 10 {
			maxErrors = 10
		}
		for i := 0; i < maxErrors; i++ {
			fmt.Printf("- %v\n", result.errors[i])
		}
		if len(result.errors) > maxErrors {
			fmt.Printf("- ... %d more errors\n", len(result.errors)-maxErrors)
		}
	}
}

func printStats(title string, values []time.Duration) {
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))
	if len(values) == 0 {
		fmt.Println("No successful samples")
		return
	}

	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	var total time.Duration
	for _, value := range values {
		total += value
	}

	avg := total / time.Duration(len(values))
	fmt.Printf("samples: %d\n", len(values))
	fmt.Printf("min:     %s\n", values[0].Round(time.Microsecond))
	fmt.Printf("avg:     %s\n", avg.Round(time.Microsecond))
	fmt.Printf("p50:     %s\n", percentile(values, 50).Round(time.Microsecond))
	fmt.Printf("p95:     %s\n", percentile(values, 95).Round(time.Microsecond))
	fmt.Printf("p99:     %s\n", percentile(values, 99).Round(time.Microsecond))
	fmt.Printf("max:     %s\n", values[len(values)-1].Round(time.Microsecond))
}

func percentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}
	position := (p / 100) * float64(len(values)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return values[lower]
	}
	weight := position - float64(lower)
	return time.Duration(float64(values[lower])*(1-weight) + float64(values[upper])*weight)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
