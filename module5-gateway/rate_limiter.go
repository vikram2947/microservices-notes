package main

import (
	"net/http"
	"sync"
	"time"
)

type tokenBucket struct {
	mu       sync.Mutex
	capacity int
	tokens   int
	fillInt  time.Duration
	lastFill time.Time
}

func newTokenBucket(capacity int, fillInt time.Duration) *tokenBucket {
	return &tokenBucket{
		capacity: capacity,
		tokens:   capacity,
		fillInt:  fillInt,
		lastFill: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastFill)
	if elapsed >= b.fillInt {
		add := int(elapsed / b.fillInt)
		b.tokens += add
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastFill = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func rateLimitMiddleware(bucket *tokenBucket, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !bucket.allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

