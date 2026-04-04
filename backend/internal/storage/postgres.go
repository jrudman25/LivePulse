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
	Location      string    `json:"location"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	ExternalAPIID string    `json:"external_api_id"` // ID from Ticketmaster/SeatGeek
	CreatedAt     time.Time `json:"created_at"`
	IsFavorite    bool      `json:"is_favorite"`     // Dynamic append flag for client payload
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

	ALTER TABLE events ADD COLUMN IF NOT EXISTS location VARCHAR(255);
	`
	_, err := db.pool.Exec(ctx, queries)
	return err
}

// InsertEvent cleanly inserts or updates an event in Postgres
func (db *PostgresClient) InsertEvent(ctx context.Context, e Event) error {
	query := `
		INSERT INTO events (id, type, title, location, start_time, end_time, external_api_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			location = EXCLUDED.location,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time;
	`
	_, err := db.pool.Exec(ctx, query, e.ID, e.Type, e.Title, e.Location, e.StartTime, e.EndTime, e.ExternalAPIID, e.CreatedAt)
	return err
}

// Close gracefully closes the database pool
func (db *PostgresClient) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// GetUpcomingEvents fetches events ordered by date
func (db *PostgresClient) GetUpcomingEvents(ctx context.Context, limit int) ([]Event, error) {
	query := `
		SELECT id, type, title, location, start_time, end_time, external_api_id, created_at
		FROM events
		WHERE end_time > NOW()
		ORDER BY start_time ASC
		LIMIT $1
	`
	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var loc *string
		if err := rows.Scan(&e.ID, &e.Type, &e.Title, &loc, &e.StartTime, &e.EndTime, &e.ExternalAPIID, &e.CreatedAt); err != nil {
			return nil, err
		}
		if loc != nil { e.Location = *loc }
		events = append(events, e)
	}
	return events, nil
}

// AddFavorite links a user to an event bookmark
func (db *PostgresClient) AddFavorite(ctx context.Context, userID, eventID string) error {
	query := `
		INSERT INTO favorites (user_id, event_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, event_id) DO NOTHING
	`
	// Upsert the user into the database as a reference since Clerk handles auth natively
	upsertUser := `INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`
	_, _ = db.pool.Exec(ctx, upsertUser, userID)

	_, err := db.pool.Exec(ctx, query, userID, eventID)
	return err
}

// RemoveFavorite unlinks a user from an event bookmark
func (db *PostgresClient) RemoveFavorite(ctx context.Context, userID, eventID string) error {
	query := `
		DELETE FROM favorites WHERE user_id = $1 AND event_id = $2
	`
	_, err := db.pool.Exec(ctx, query, userID, eventID)
	return err
}

// GetUserFavorites fetches all favorites for a specific user
func (db *PostgresClient) GetUserFavorites(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT event_id FROM favorites WHERE user_id = $1`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		eventIDs = append(eventIDs, id)
	}
	return eventIDs, nil
}
