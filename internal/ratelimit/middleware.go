package ratelimit

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// HTTPMiddleware returns an HTTP middleware that rate-limits requests per job ID.
// It expects the job ID to be the last path segment, e.g. /jobs/{id}/start.
// Requests that are suppressed receive a 429 Too Many Requests response.
func HTTPMiddleware(rl *RateLimiter, cooldown time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobID := extractLastSegment(r.URL.Path)
		if jobID == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !rl.Allow(jobID, cooldown) {
			log.Printf("[ratelimit] suppressed alert for job %q (cooldown %s)", jobID, cooldown)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited","message":"alert suppressed during cooldown period"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractLastSegment returns the last non-empty path segment.
func extractLastSegment(path string) string {
	path = strings.TrimRight(path, "/")
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return path
	}
	return path[idx+1:]
}
