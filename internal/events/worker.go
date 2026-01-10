package events

import (
	"context"
	"log"
	"sync"
)

// EventHandler is a function that processes an event
type EventHandler func(*Event) error

// WorkerPool manages a pool of worker goroutines that process events
type WorkerPool struct {
	queue       *Queue
	workerCount int
	handler     EventHandler
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(queue *Queue, workerCount int, handler EventHandler) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		queue:       queue,
		workerCount: workerCount,
		handler:     handler,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start launches all worker goroutines
func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workerCount)
	
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker is the main loop for each worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	log.Printf("Worker %d started", id)
	
	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		default:
			event, ok := wp.queue.Dequeue(wp.ctx)
			if !ok {
				// Queue is closed or context cancelled
				log.Printf("Worker %d: queue closed or context cancelled", id)
				return
			}
			
			if event == nil {
				continue
			}
			
			// Process the event
			if err := wp.handler(event); err != nil {
				log.Printf("Worker %d: error processing event %s: %v", id, event.ID, err)
			}
		}
	}
}

// Shutdown gracefully shuts down the worker pool
// It waits for all workers to finish processing their current events
func (wp *WorkerPool) Shutdown() {
	log.Println("Shutting down worker pool...")
	
	// Signal all workers to stop
	wp.cancel()
	
	// Wait for all workers to finish
	wp.wg.Wait()
	
	log.Println("Worker pool shutdown complete")
}

// ShutdownWithDrain gracefully shuts down and processes remaining events
func (wp *WorkerPool) ShutdownWithDrain() {
	log.Println("Shutting down worker pool with drain...")
	
	// Close the queue to prevent new events
	wp.queue.Close()
	
	// Process remaining events
	remaining := wp.queue.Drain()
	log.Printf("Processing %d remaining events", len(remaining))
	
	for _, event := range remaining {
		if err := wp.handler(event); err != nil {
			log.Printf("Error processing remaining event %s: %v", event.ID, err)
		}
	}
	
	// Cancel context and wait for workers
	wp.cancel()
	wp.wg.Wait()
	
	log.Println("Worker pool shutdown with drain complete")
}
