package middleware

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestApplicationRateLimiterDisablesLimitsOnlyForE2E(t *testing.T) {
	t.Parallel()

	limiter := newApplicationRateLimiter("e2e", rate.Every(time.Hour), 1)
	for range 10 {
		if !limiter.get("127.0.0.1").Allow() {
			t.Fatal("expected the isolated E2E limiter to allow every request")
		}
	}
}

func TestApplicationRateLimiterKeepsLimitsOutsideE2E(t *testing.T) {
	t.Parallel()

	limiter := newApplicationRateLimiter("production", rate.Every(time.Hour), 1)
	if !limiter.get("127.0.0.1").Allow() {
		t.Fatal("expected first request to be allowed")
	}
	if limiter.get("127.0.0.1").Allow() {
		t.Fatal("expected rate limit outside the E2E environment")
	}
}
