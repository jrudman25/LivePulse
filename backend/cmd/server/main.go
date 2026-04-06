package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jrudman25/livepulse/config"
	"github.com/jrudman25/livepulse/internal/aggregation"
	"github.com/jrudman25/livepulse/internal/api"
	"github.com/jrudman25/livepulse/internal/events"
	"github.com/jrudman25/livepulse/internal/milestones"
	"github.com/jrudman25/livepulse/internal/storage"
	"github.com/TwiN/go-away"
)

func main() {
	log.Println("Starting LivePulse...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Configuration loaded: %d workers, queue size %d", cfg.Worker.Count, cfg.Worker.EventQueueSize)

	// Initialize Clerk Auth
	clerkSecret := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecret != "" {
		api.SetClerkKey(clerkSecret)
	}

	// Initialize Postgres
	pgClient, err := storage.NewPostgresClient(context.Background(), cfg.Postgres.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer pgClient.Close()

	if err := pgClient.InitSchema(context.Background()); err != nil {
		log.Fatalf("Failed to init postgres schema: %v", err)
	}
	log.Println("Postgres initialized")

	// Initialize Redis
	redisClient, err := storage.NewRedisClient(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis initialized")

	// Initialize API Fetcher (Cron)
	apiFetcher := events.NewAPIFetcher(pgClient, os.Getenv("EXTERNAL_API_KEY"))
	apiFetcher.Start()
	defer apiFetcher.Stop()

	// Create event queue
	eventQueue := events.NewQueue(cfg.Worker.EventQueueSize)
	log.Printf("Event queue created with size %d", cfg.Worker.EventQueueSize)

	// Create aggregation manager
	aggManager := aggregation.NewManager()
	log.Println("Aggregation manager initialized")

	// Create WebSocket hub
	wsHub := api.NewWebSocketHub()
	log.Println("WebSocket hub initialized")

	// Create milestone tracker with notification handler
	tracker := milestones.NewTracker(func(achievement *milestones.MilestoneAchievement) {
		log.Printf("MILESTONE ACHIEVED: %s - %s", achievement.SessionID, achievement.Milestone.Description)

		// Broadcast to WebSocket clients
		wsHub.BroadcastToSession(achievement.SessionID, map[string]interface{}{
			"type":        "milestone_achieved",
			"milestone":   achievement.Milestone,
			"achieved_at": achievement.AchievedAt,
		})
	})
	log.Println("Milestone tracker initialized")

	// Create event handler
	eventHandler := func(event *events.Event) error {
		// Update aggregation
		aggManager.ProcessEvent(event)

		// Check milestones
		if stats, exists := aggManager.GetSession(event.SessionID); exists {
			tracker.CheckMilestones(event.SessionID, stats)
		}

		// Broadcast event to WebSocket clients
		if event.Type == events.EventTypeReaction {
			if reactionType, ok := event.GetReactionType(); ok {
				wsHub.BroadcastToSession(event.SessionID, map[string]interface{}{
					"type":          "reaction",
					"user_id":       event.UserID,
					"reaction_type": reactionType,
					"timestamp":     event.Timestamp,
				})
			}
		} else if event.Type == events.EventTypeChat {
			if text, authorName, ok := event.GetChatText(); ok {
				// Censor profanity using go-away
				cleanText := goaway.Censor(text)

				chatMsg := &storage.ChatMessage{
					ID:         event.ID,
					UserID:     event.UserID,
					SessionID:  event.SessionID,
					Text:       cleanText,
					AuthorName: authorName,
					Timestamp:  event.Timestamp,
				}
				
				// Save to Redis
				if err := redisClient.SaveChatMessage(context.Background(), event.SessionID, chatMsg); err != nil {
					log.Printf("Error saving chat message to redis: %v", err)
				}

				// Broadcast
				wsHub.BroadcastToSession(event.SessionID, map[string]interface{}{
					"type":    "chat",
					"message": chatMsg,
				})
			}
		}

		return nil
	}

	// Create and start worker pool
	workerPool := events.NewWorkerPool(eventQueue, cfg.Worker.Count, eventHandler)
	workerPool.Start()
	log.Printf("Worker pool started with %d workers", cfg.Worker.Count)

	// Create API server
	apiServer := api.NewServer(eventQueue, aggManager, tracker, wsHub, pgClient)

	// Set up HTTP routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", api.Chain(apiServer.HandleHealth, api.LoggingMiddleware, api.CORSMiddleware))

	// Session management
	mux.HandleFunc("/api/sessions", api.Chain(apiServer.HandleCreateSession, api.LoggingMiddleware, api.CORSMiddleware, api.RecoveryMiddleware))
	mux.HandleFunc("/api/sessions/join", api.Chain(apiServer.HandleJoinSession, api.LoggingMiddleware, api.CORSMiddleware, api.RecoveryMiddleware))
	mux.HandleFunc("/api/sessions/stats", api.Chain(apiServer.HandleGetStats, api.LoggingMiddleware, api.CORSMiddleware, api.RecoveryMiddleware))
	mux.HandleFunc("/api/sessions/milestones", api.Chain(apiServer.HandleGetMilestones, api.LoggingMiddleware, api.CORSMiddleware, api.RecoveryMiddleware))

	// API integration routes
	mux.HandleFunc("/api/events", api.Chain(apiServer.HandleGetLiveEvents, api.LoggingMiddleware, api.CORSMiddleware))
	mux.HandleFunc("/api/events/single", api.Chain(apiServer.HandleGetEvent, api.LoggingMiddleware, api.CORSMiddleware))
	mux.HandleFunc("/api/favorites", api.Chain(apiServer.HandleToggleFavorite, api.LoggingMiddleware, api.CORSMiddleware, api.ClerkMiddleware))
	
	// Admin trigger for Ticketmaster
	mux.HandleFunc("/api/admin/trigger-fetch", api.Chain(func(w http.ResponseWriter, r *http.Request) {
		go apiFetcher.FetchAPIEvents()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ticketmaster fetch triggered"}`))
	}, api.LoggingMiddleware, api.CORSMiddleware))

	// WebSocket
	mux.HandleFunc("/ws", apiServer.HandleWebSocket)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("HTTP server listening on :%s", cfg.Server.Port)
		log.Printf("WebSocket endpoint: ws://localhost:%s/ws", cfg.Server.Port)
		log.Printf("API endpoint: http://localhost:%s/api", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\nShutting down...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("HTTP server stopped")

	// Shutdown worker pool with drain
	workerPool.ShutdownWithDrain()
	log.Println("Worker pool stopped")

	log.Println("LivePulse shutdown complete. Goodbye!")
}
