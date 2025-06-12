package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter represents a client-specific rate limiter
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware creates a rate limiting middleware
type RateLimitMiddleware struct {
	clients map[string]*RateLimiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(rps float64, burst int) *RateLimitMiddleware {
	rl := &RateLimitMiddleware{
		clients: make(map[string]*RateLimiter),
		rate:    rate.Limit(rps),
		burst:   burst,
	}

	// Clean up old clients every minute
	go rl.cleanupRoutine()

	return rl
}

// getLimiter gets or creates a rate limiter for a client
func (rl *RateLimitMiddleware) getLimiter(clientIP string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clients[clientIP]
	if !exists {
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.clients[clientIP] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter
}

// cleanupRoutine removes old clients to prevent memory leaks
func (rl *RateLimitMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		rl.mu.Lock()
		for clientIP, limiter := range rl.clients {
			if time.Since(limiter.lastSeen) > 3*time.Minute {
				delete(rl.clients, clientIP)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns the Gin middleware function
func (rl *RateLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := rl.getLimiter(clientIP)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
