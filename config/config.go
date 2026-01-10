package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Worker    WorkerConfig
	DynamoDB  DynamoDBConfig
	Milestone MilestoneConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// WorkerConfig holds worker pool configuration
type WorkerConfig struct {
	Count          int
	EventQueueSize int
}

// DynamoDBConfig holds DynamoDB connection configuration
type DynamoDBConfig struct {
	Endpoint        string
	Region          string
	EventsTable     string
	SessionsTable   string
	MilestonesTable string
}

// MilestoneConfig holds milestone tracking configuration
type MilestoneConfig struct {
	Thresholds []int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "15s")),
			WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "15s")),
		},
		Worker: WorkerConfig{
			Count:          parseInt(getEnv("WORKER_COUNT", "10")),
			EventQueueSize: parseInt(getEnv("EVENT_QUEUE_SIZE", "10000")),
		},
		DynamoDB: DynamoDBConfig{
			Endpoint:        getEnv("DYNAMODB_ENDPOINT", ""),
			Region:          getEnv("AWS_REGION", "us-west-2"),
			EventsTable:     getEnv("DYNAMODB_EVENTS_TABLE", "LivePulse_Events"),
			SessionsTable:   getEnv("DYNAMODB_SESSIONS_TABLE", "LivePulse_Sessions"),
			MilestonesTable: getEnv("DYNAMODB_MILESTONES_TABLE", "LivePulse_Milestones"),
		},
		Milestone: MilestoneConfig{
			Thresholds: parseIntSlice(getEnv("MILESTONE_THRESHOLDS", "100,500,1000,5000,10000")),
		},
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseInt parses a string to int
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// parseDuration parses a string to time.Duration
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Second
	}
	return d
}

// parseIntSlice parses a comma-separated string to []int
func parseIntSlice(s string) []int {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if val, err := strconv.Atoi(part); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Worker.Count <= 0 {
		return fmt.Errorf("worker count must be positive")
	}
	if c.Worker.EventQueueSize <= 0 {
		return fmt.Errorf("event queue size must be positive")
	}
	if c.DynamoDB.Region == "" {
		return fmt.Errorf("AWS region is required")
	}
	return nil
}
