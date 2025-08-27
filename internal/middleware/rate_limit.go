package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/gin-gonic/gin"
)

// RateLimiter represents a rate limiter
type RateLimiter struct {
	clients map[string]*ClientLimiter
	mutex   sync.RWMutex
	limit   int
	window  time.Duration
}

// ClientLimiter represents a client-specific rate limiter
type ClientLimiter struct {
	tokens    int
	lastReset time.Time
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		limit:   limit,
		window:  window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mutex.RLock()
	client, exists := rl.clients[clientID]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		// Double-check after acquiring write lock
		if client, exists = rl.clients[clientID]; !exists {
			client = &ClientLimiter{
				tokens:    rl.limit,
				lastReset: time.Now(),
			}
			rl.clients[clientID] = client
		}
		rl.mutex.Unlock()
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	now := time.Now()
	// Reset tokens if window has passed
	if now.Sub(client.lastReset) >= rl.window {
		client.tokens = rl.limit
		client.lastReset = now
	}

	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

// cleanup removes old clients periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for clientID, client := range rl.clients {
			client.mutex.Lock()
			if now.Sub(client.lastReset) > rl.window*2 {
				delete(rl.clients, clientID)
			}
			client.mutex.Unlock()
		}
		rl.mutex.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP address as client identifier
		clientIP := c.ClientIP()

		// Check for authentication and use user ID if available
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(uint); ok {
				clientIP = string(rune(id)) // Use user ID for authenticated requests
			}
		}

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponseWithCodeAndError(
				http.StatusTooManyRequests,
				"Rate limit exceeded",
				"RATE_LIMIT_EXCEEDED",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreateRateLimiters creates different rate limiters for different endpoints
func CreateRateLimiters() map[string]*RateLimiter {
	return map[string]*RateLimiter{
		"auth":   NewRateLimiter(5, time.Minute),     // 5 requests per minute for auth endpoints
		"admin":  NewRateLimiter(100, time.Minute),   // 100 requests per minute for admin endpoints
		"client": NewRateLimiter(1000, time.Minute),  // 1000 requests per minute for client endpoints
		"global": NewRateLimiter(10000, time.Minute), // 10000 requests per minute globally
	}
}

// RateLimitByType creates a rate limiting middleware for specific endpoint types
func RateLimitByType(limiters map[string]*RateLimiter, limiterType string) gin.HandlerFunc {
	limiter, exists := limiters[limiterType]
	if !exists {
		// Fallback to global limiter
		limiter = limiters["global"]
	}

	return RateLimitMiddleware(limiter)
}
