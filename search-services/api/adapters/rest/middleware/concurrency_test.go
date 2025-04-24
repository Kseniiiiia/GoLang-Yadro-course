package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConcurrencyMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		parallel      int
		expectedCodes []int
	}{
		{
			name:          "single request",
			limit:         1,
			parallel:      1,
			expectedCodes: []int{http.StatusOK},
		},
		{
			name:          "limit not exceeded",
			limit:         2,
			parallel:      2,
			expectedCodes: []int{http.StatusOK, http.StatusOK},
		},
		{
			name:          "limit exceeded",
			limit:         1,
			parallel:      2,
			expectedCodes: []int{http.StatusOK, http.StatusServiceUnavailable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Concurrency(
				func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond) // Имитация обработки
					w.WriteHeader(http.StatusOK)
				},
				tt.limit,
			)

			results := make(chan int, tt.parallel)
			for i := 0; i < tt.parallel; i++ {
				go func() {
					rr := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/", nil)
					handler.ServeHTTP(rr, req)
					results <- rr.Code
				}()
			}

			var gotCodes []int
			for i := 0; i < tt.parallel; i++ {
				gotCodes = append(gotCodes, <-results)
			}

			if len(gotCodes) != len(tt.expectedCodes) {
				t.Fatalf("expected %d responses, got %d", len(tt.expectedCodes), len(gotCodes))
			}

			hasOK := false
			has503 := false
			for _, code := range gotCodes {
				if code == http.StatusOK {
					hasOK = true
				}
				if code == http.StatusServiceUnavailable {
					has503 = true
				}
			}

			if tt.limit < tt.parallel && !(hasOK && has503) {
				t.Errorf("expected both OK and 503 statuses when limit exceeded")
			}
		})
	}
}
