package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := NewQueue(10)
	defer q.Close()

	event := ChatEvent("session-1", "user-1", "hello", "Jordan")
	ok := q.Enqueue(event)
	assert.True(t, ok, "enqueue should succeed on empty queue")
	assert.Equal(t, 1, q.Len())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out, ok := q.Dequeue(ctx)
	require.True(t, ok)
	assert.Equal(t, event.ID, out.ID)
	assert.Equal(t, 0, q.Len())
}

func TestQueue_DropsWhenFull(t *testing.T) {
	q := NewQueue(2)
	defer q.Close()

	e1 := ChatEvent("s", "u", "msg1", "A")
	e2 := ChatEvent("s", "u", "msg2", "A")
	e3 := ChatEvent("s", "u", "msg3", "A")

	assert.True(t, q.Enqueue(e1))
	assert.True(t, q.Enqueue(e2))
	assert.False(t, q.Enqueue(e3), "third enqueue should fail on a queue of capacity 2")
	assert.Equal(t, 2, q.Len())
}

func TestQueue_RejectsAfterClose(t *testing.T) {
	q := NewQueue(10)
	q.Close()

	event := ChatEvent("s", "u", "msg", "A")
	assert.False(t, q.Enqueue(event), "enqueue should fail on closed queue")
	assert.True(t, q.IsClosed())
}

func TestQueue_DrainReturnsRemaining(t *testing.T) {
	q := NewQueue(10)

	q.Enqueue(ChatEvent("s", "u", "msg1", "A"))
	q.Enqueue(ChatEvent("s", "u", "msg2", "A"))
	q.Enqueue(ChatEvent("s", "u", "msg3", "A"))

	q.Close()
	remaining := q.Drain()
	assert.Len(t, remaining, 3)
}

func TestQueue_DequeueRespectsContextCancellation(t *testing.T) {
	q := NewQueue(10)
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, ok := q.Dequeue(ctx)
	assert.False(t, ok, "dequeue should return false when context expires on empty queue")
}

func TestQueue_ConcurrentEnqueue(t *testing.T) {
	q := NewQueue(1000)
	defer q.Close()

	var wg sync.WaitGroup
	successes := make(chan bool, 500)

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := ChatEvent("s", "u", "msg", "A")
			successes <- q.Enqueue(e)
		}()
	}
	wg.Wait()
	close(successes)

	count := 0
	for ok := range successes {
		if ok {
			count++
		}
	}
	assert.Equal(t, 500, count, "all 500 enqueues should succeed with capacity 1000")
	assert.Equal(t, 500, q.Len())
}

func TestQueue_CapReturnsBufferSize(t *testing.T) {
	q := NewQueue(42)
	defer q.Close()
	assert.Equal(t, 42, q.Cap())
}
