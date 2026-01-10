package events

import (
	"context"
	"log"
	"sync"
)

// Queue manages the event queue using a buffered channel
type Queue struct {
	events   chan *Event
	size     int
	mu       sync.RWMutex
	closed   bool
	draining bool
}

// NewQueue creates a new event queue with the specified buffer size
func NewQueue(size int) *Queue {
	return &Queue{
		events: make(chan *Event, size),
		size:   size,
	}
}

// Enqueue adds an event to the queue
// Returns false if the queue is full or closed
func (q *Queue) Enqueue(event *Event) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return false
	}

	select {
	case q.events <- event:
		return true
	default:
		// Queue is full, event is dropped
		log.Printf("WARNING: Event queue full, dropping event %s", event.ID)
		return false
	}
}

// Dequeue retrieves the next event from the queue
// Returns nil if the queue is closed and empty
func (q *Queue) Dequeue(ctx context.Context) (*Event, bool) {
	select {
	case event, ok := <-q.events:
		return event, ok
	case <-ctx.Done():
		return nil, false
	}
}

// Close closes the queue and prevents new events from being enqueued
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.closed {
		q.closed = true
		close(q.events)
	}
}

// Drain processes all remaining events in the queue
func (q *Queue) Drain() []*Event {
	q.mu.Lock()
	q.draining = true
	q.mu.Unlock()

	var remaining []*Event
	for event := range q.events {
		remaining = append(remaining, event)
	}
	return remaining
}

// Len returns the current number of events in the queue
func (q *Queue) Len() int {
	return len(q.events)
}

// Cap returns the capacity of the queue
func (q *Queue) Cap() int {
	return q.size
}

// IsClosed returns whether the queue is closed
func (q *Queue) IsClosed() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.closed
}
