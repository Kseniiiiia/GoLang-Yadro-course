package middleware

import (
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	rate          int
	capacity      int
	tokens        int
	lastTimestamp time.Time
	mu            sync.Mutex
}

func NewTokenBucket(rate int) *TokenBucket {
	return &TokenBucket{
		rate:          rate,
		capacity:      rate,
		tokens:        rate,
		lastTimestamp: time.Now(),
	}
}

func (tb *TokenBucket) Wait() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTimestamp)
	tb.lastTimestamp = now

	newTokens := int(elapsed.Seconds() * float64(tb.rate))
	if newTokens > 0 {
		tb.tokens = min(tb.tokens+newTokens, tb.capacity)
	}

	if tb.tokens > 0 {
		tb.tokens--
		return
	}

	waitTime := time.Duration(float64(time.Second) / float64(tb.rate))
	time.Sleep(waitTime)
	tb.lastTimestamp = time.Now()
}

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	limiter := NewTokenBucket(rps)

	return func(w http.ResponseWriter, r *http.Request) {
		limiter.Wait()
		next(w, r)
	}
}
