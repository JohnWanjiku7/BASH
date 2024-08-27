package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter controls the rate at which requests are allowed
type RateLimiter struct {
	mu                 sync.Mutex
	userRequestTimes   map[string][]time.Time
	requestWindow      time.Duration
	maxAllowedRequests int
}

// NewRateLimiter creates a new RateLimiter instance with specified settings
func NewRateLimiter(requestWindow time.Duration, maxAllowedRequests int) *RateLimiter {
	return &RateLimiter{
		userRequestTimes:   make(map[string][]time.Time),
		requestWindow:      requestWindow,
		maxAllowedRequests: maxAllowedRequests,
	}
}

// Limit returns a Gin middleware that limits the number of requests per user
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the user ID from the request context
		userID, userIDExists := c.Get("user_id")
		if !userIDExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			//c.Abort()
			return
		}

		userIDStr, userIDValid := userID.(string)
		if !userIDValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		rl.mu.Lock()
		defer rl.mu.Unlock()

		currentTime := time.Now()
		requestTimes, exists := rl.userRequestTimes[userIDStr]

		if !exists {
			// No previous request records for this user
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
		rl.userRequestTimes[userIDStr] = requestTimes

		c.Next()
	}
}
