package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter controls the rate at which requests are allowed
type IPRateLimiter struct {
	mu                 sync.Mutex
	ipRequestTimes     map[string][]time.Time
	requestWindow      time.Duration
	maxAllowedRequests int
}

// NewRateLimiter creates a new RateLimiter instance with specified settings
func NewIPRateLimiter(requestWindow time.Duration, maxAllowedRequests int) *IPRateLimiter {
	return &IPRateLimiter{
		ipRequestTimes:     make(map[string][]time.Time),
		requestWindow:      requestWindow,
		maxAllowedRequests: maxAllowedRequests,
	}
}

// Limit returns a Gin middleware that limits the number of requests per IP address
func (rl *IPRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the client IP address
		clientIP := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		currentTime := time.Now()
		requestTimes, exists := rl.ipRequestTimes[clientIP]

		if !exists {
			// No previous request records for this IP
			requestTimes = []time.Time{}
		} else {
			// Filter out timestamps outside the request window
			validRequestTimes := []time.Time{}
			for _, timestamp := range requestTimes {
				if currentTime.Sub(timestamp) < rl.requestWindow {
					validRequestTimes = append(validRequestTimes, timestamp)
				}
			}
			requestTimes = validRequestTimes
		}

		if len(requestTimes) >= rl.maxAllowedRequests {
			// Rate limit exceeded
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			c.Abort()
			return
		}

		// Record the current request time
		requestTimes = append(requestTimes, currentTime)
		rl.ipRequestTimes[clientIP] = requestTimes

		c.Next()
	}
}
