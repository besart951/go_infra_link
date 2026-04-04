package middleware

import (
	"net/http"
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

// LoginRateLimit allows 5 login attempts per 30 seconds (burst 5) per IP.
// This is applied only to POST /api/v1/auth/login.
var loginLimiter = newLoginRateLimiter(rate.Every(6*time.Second), 5)

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
