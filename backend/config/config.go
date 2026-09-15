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
	Environment string // "development" (default) or "production"
	Server      ServerConfig
	Worker      WorkerConfig
	Milestone   MilestoneConfig
	Postgres    PostgresConfig
	Redis       RedisConfig
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

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	DatabaseURL string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	URL string
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
		Environment: strings.ToLower(getEnv("APP_ENV", "development")),
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "15s")),
			WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "15s")),
		},
		Worker: WorkerConfig{
			Count:          parseInt(getEnv("WORKER_COUNT", "10")),
			EventQueueSize: parseInt(getEnv("EVENT_QUEUE_SIZE", "10000")),
		},
		Postgres: PostgresConfig{
			DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/livepulse?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
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
	if c.Postgres.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}

	// In production, fail fast when the secrets and transport security the
	// service actually needs are missing rather than reporting healthy.
	if c.Environment == "production" {
		if os.Getenv("CLERK_SECRET_KEY") == "" {
			return fmt.Errorf("CLERK_SECRET_KEY is required in production")
		}
		if apiKey := os.Getenv("EXTERNAL_API_KEY"); apiKey == "" || apiKey == "your_ticketmaster_api_key" {
			return fmt.Errorf("EXTERNAL_API_KEY is required in production")
		}
		if strings.Contains(c.Postgres.DatabaseURL, "sslmode=disable") {
			return fmt.Errorf("DATABASE_URL must use TLS in production (sslmode=disable is not allowed)")
		}
		if !strings.HasPrefix(c.Redis.URL, "rediss://") {
			return fmt.Errorf("REDIS_URL must use TLS in production (rediss://)")
		}
		if os.Getenv("CORS_ALLOWED_ORIGINS") == "" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS is required in production")
		}
		if os.Getenv("WS_ALLOWED_ORIGINS") == "" {
			return fmt.Errorf("WS_ALLOWED_ORIGINS is required in production")
		}
	}
	return nil
}
