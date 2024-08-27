package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func RequestUniqueId() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		c.Next()
	}
}

func MultiTenantRouting() gin.HandlerFunc {
	return func(c *gin.Context) {

		restaurantIDStr := c.Param("restaurantId")
		restaurantUUID, err := uuid.Parse(restaurantIDStr)
		if err != nil {
			log.Warn().Msg("Restaurant is missing from the URL")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID"})
			c.Abort()
			return
		}
		c.Set("restaurant", restaurantUUID.String())
		c.Next()
	}
}
