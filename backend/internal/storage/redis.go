package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client
type RedisClient struct {
	client *redis.Client
}

// ChatMessage represents a single message in the live chat
type ChatMessage struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	SessionID  string    `json:"session_id"` // Matches Event ID
	Text       string    `json:"text"`
	AuthorName string    `json:"author_name"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewRedisClient establishes a connection to Upstash/Redis
func NewRedisClient(url string) (*RedisClient, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// SaveChatMessage adds a message to the event's chat list and ensures a TTL is set
func (rc *RedisClient) SaveChatMessage(ctx context.Context, sessionID string, msg *ChatMessage) error {
	key := fmt.Sprintf("chat:%s", sessionID)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Add message and enforce the 500-message retention cap atomically;
	// a failed LTRIM must surface as an error rather than silently
	// letting the list grow past the documented limit.
	pipe := rc.client.TxPipeline()
	pipe.RPush(ctx, key, data)
	pipe.LTrim(ctx, key, -500, -1)
	_, err = pipe.Exec(ctx)
	return err
}

// GetRecentChat fetches the chat history for an event
func (rc *RedisClient) GetRecentChat(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	key := fmt.Sprintf("chat:%s", sessionID)

	// Fetch all messages (up to 500 based on LTrim)
	results, err := rc.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var messages []ChatMessage
	for _, res := range results {
		var msg ChatMessage
		if err := json.Unmarshal([]byte(res), &msg); err == nil {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// SetChatTTL configures the chat key to expire some time after the event ends
func (rc *RedisClient) SetChatTTL(ctx context.Context, sessionID string, expireAt time.Time) error {
	key := fmt.Sprintf("chat:%s", sessionID)
	// We add 1 hour to the event's end time per the feature requirements
	deletionTime := expireAt.Add(1 * time.Hour)

	// ExpireAt explicitly schedules the key for deletion at a specific time
	return rc.client.ExpireAt(ctx, key, deletionTime).Err()
}

// Ping verifies Redis connectivity for readiness checks
func (rc *RedisClient) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close gracefully closes the redis client
func (rc *RedisClient) Close() error {
	if rc.client != nil {
		return rc.client.Close()
	}
	return nil
}
