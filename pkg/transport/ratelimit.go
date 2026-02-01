package transport

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	tokens         chan struct{}
	refillRate     time.Duration
	capacity       int
	stopRefill     chan struct{}
	refillStopped  chan struct{}
	mu             sync.Mutex
	started        bool
	stopOnce       sync.Once
}

// NewRateLimiter creates a new rate limiter with the specified requests per second.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10 // Default to 10 requests per second
	}

	rl := &RateLimiter{
		tokens:        make(chan struct{}, requestsPerSecond),
		capacity:      requestsPerSecond,
		refillRate:    time.Second / time.Duration(requestsPerSecond),
		stopRefill:    make(chan struct{}),
		refillStopped: make(chan struct{}),
	}

	// Fill initial tokens
	for i := 0; i < requestsPerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	return rl
}

// Start begins the token refill process.
func (rl *RateLimiter) Start() {
	rl.mu.Lock()
	if rl.started {
		rl.mu.Unlock()
		return
	}
	rl.started = true
	rl.mu.Unlock()

	go rl.refill()
}

// Wait blocks until a token is available or the context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false otherwise.
func (rl *RateLimiter) TryAcquire() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Stop stops the token refill process.
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	if !rl.started {
		rl.mu.Unlock()
		return
	}
	rl.mu.Unlock()

	rl.stopOnce.Do(func() {
		close(rl.stopRefill)
	})
	<-rl.refillStopped
}

// refill continuously adds tokens to the bucket at the specified rate.
func (rl *RateLimiter) refill() {
	defer close(rl.refillStopped)

	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopRefill:
			return
		case <-ticker.C:
			// Try to add a token, but don't block if the bucket is full
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Bucket is full, skip this refill
			}
		}
	}
}

// Capacity returns the maximum number of tokens the bucket can hold.
func (rl *RateLimiter) Capacity() int {
	return rl.capacity
}

// Available returns the approximate number of tokens currently available.
// This is an estimate and may not be exact due to concurrent access.
func (rl *RateLimiter) Available() int {
	return len(rl.tokens)
}
