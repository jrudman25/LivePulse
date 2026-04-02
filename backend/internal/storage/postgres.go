package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresClient wraps the pgx database pool
type PostgresClient struct {
	pool *pgxpool.Pool
}

// User represents an authenticated user from Clerk
type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// Event represents a scheduled live event
type Event struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`            // e.g., "concert", "sports"
	Title         string    `json:"title"`           // User friendly name
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	ExternalAPIID string    `json:"external_api_id"` // ID from Ticketmaster/SeatGeek
	CreatedAt     time.Time `json:"created_at"`
}

// Favorite represents a user's bookmarked event
type Favorite struct {
	UserID    string    `json:"user_id"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewPostgresClient establishes a connection pool to Neon/PostgreSQL
func NewPostgresClient(ctx context.Context, databaseURL string) (*PostgresClient, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	// Optimize for Serverless Postgres
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PostgresClient{pool: pool}, nil
}

// InitSchema sets up the required tables if they don't exist
func (db *PostgresClient) InitSchema(ctx context.Context) error {
	queries := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(255) PRIMARY KEY,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id VARCHAR(255) PRIMARY KEY,
		type VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		start_time TIMESTAMP WITH TIME ZONE NOT NULL,
		end_time TIMESTAMP WITH TIME ZONE NOT NULL,
		external_api_id VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS favorites (
		user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
		event_id VARCHAR(255) REFERENCES events(id) ON DELETE CASCADE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, event_id)
	);
	`
	_, err := db.pool.Exec(ctx, queries)
	return err
}

// InsertEvent cleanly inserts or updates an event in Postgres
func (db *PostgresClient) InsertEvent(ctx context.Context, e Event) error {
	query := `
		INSERT INTO events (id, type, title, start_time, end_time, external_api_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time;
	`
	_, err := db.pool.Exec(ctx, query, e.ID, e.Type, e.Title, e.StartTime, e.EndTime, e.ExternalAPIID, e.CreatedAt)
	return err
}

// Close gracefully closes the database pool
func (db *PostgresClient) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}
