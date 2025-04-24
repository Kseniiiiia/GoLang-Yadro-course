package middleware

import (
	"log/slog"
	"net/http"
)

func Concurrency(next http.HandlerFunc, limit int) http.HandlerFunc {
	var (
		semaphore = make(chan struct{}, limit)
		logger    = slog.Default()
		release   = func() { <-semaphore }
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case semaphore <- struct{}{}:
			defer release()
			next.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)

			if _, err := w.Write([]byte("Too many concurrent requests")); err != nil {
				logger.Error("failed to write concurrency limit response",
					"error", err)
			}
		}
	})
}
