package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// loginRateLimiter holds per-IP token bucket limiters.
type loginRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*entry
	// rate: max requests per second; burst: max burst size
	r rate.Limit
	b int
}

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newLoginRateLimiter creates a limiter allowing r events/s with burst b per IP.
func newLoginRateLimiter(r rate.Limit, b int) *loginRateLimiter {
	l := &loginRateLimiter{
		limiters: make(map[string]*entry),
		r:        r,
		b:        b,
	}
	// Purge stale entries every 10 minutes.
	go l.cleanup(10 * time.Minute)
	return l
}

func (l *loginRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.limiters[ip]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(l.r, l.b)}
		l.limiters[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

func (l *loginRateLimiter) cleanup(interval time.Duration) {
	for {
		time.Sleep(interval)
		l.mu.Lock()
		for ip, e := range l.limiters {
			if time.Since(e.lastSeen) > interval {
				delete(l.limiters, ip)
			}
		}
		l.mu.Unlock()
	}
}

// newApplicationRateLimiter keeps production rate limits active while allowing
// the isolated E2E stack to create independent browser sessions without all of
// them sharing Caddy's single backend-facing IP address. APP_ENV=e2e is only
// configured by docker-compose.e2e.yml and must never be used for deployment.
func newApplicationRateLimiter(appEnv string, r rate.Limit, b int) *loginRateLimiter {
	if strings.EqualFold(strings.TrimSpace(appEnv), "e2e") {
		return newLoginRateLimiter(rate.Inf, 0)
	}
	return newLoginRateLimiter(r, b)
}

// LoginRateLimit allows 5 login attempts per 30 seconds (burst 5) per IP.
// This is applied only to POST /api/v1/auth/login.
var loginLimiter = newApplicationRateLimiter(os.Getenv("APP_ENV"), rate.Every(6*time.Second), 5)

// AuthSensitiveRateLimit allows short bursts for token refresh/logout endpoints.
var authSensitiveLimiter = newApplicationRateLimiter(os.Getenv("APP_ENV"), rate.Every(3*time.Second), 10)

// RegistrationRateLimit allows moderate bursts for public invitation-link checks.
var registrationLimiter = newApplicationRateLimiter(os.Getenv("APP_ENV"), rate.Every(2*time.Second), 20)

// LoginRateLimitMiddleware rejects excessive login attempts with 429.
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !loginLimiter.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too_many_requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func AuthSensitiveRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !authSensitiveLimiter.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too_many_requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RegistrationRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !registrationLimiter.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too_many_requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
