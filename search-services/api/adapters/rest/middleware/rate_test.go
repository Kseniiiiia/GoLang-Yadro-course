package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		rps      int
		requests int
		minTime  time.Duration
	}{
		{
			name:     "1 rps - 1 request",
			rps:      1,
			requests: 1,
			minTime:  0,
		},
		{
			name:     "1 rps - 2 requests",
			rps:      1,
			requests: 2,
			minTime:  100 * time.Millisecond, // Должно занять ~1s между запросами
		},
		{
			name:     "10 rps - 11 requests",
			rps:      10,
			requests: 11,
			minTime:  100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Rate(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
				tt.rps,
			)

			start := time.Now()
			for i := 0; i < tt.requests; i++ {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/", nil)
				handler.ServeHTTP(rr, req)
			}
			elapsed := time.Since(start)

			if elapsed < tt.minTime {
				t.Errorf("expected at least %v to process %d requests at %d rps, took %v",
					tt.minTime, tt.requests, tt.rps, elapsed)
			}
		})
	}
}

func TestTokenBucket(t *testing.T) {
	tb := NewTokenBucket(10)

	start := time.Now()
	for i := 0; i < 11; i++ {
		tb.Wait()
	}
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Errorf("expected at least 900ms to process 11 requests at 10 rps, took %v", elapsed)
	}
}
