package middleware

import (
	"net/http"
	"strings"

	"the-dancing-pony-v2-lcwqre/helper"
	"the-dancing-pony-v2-lcwqre/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var jwtSecret = []byte(helper.GetJWTSecret())

// AuthMiddleware validates JWT tokens and ensures the user is authenticated.
func AuthMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the restaurantId from the URL
		restaurantIDStr := c.Param("restaurantId")

		// Extract token from Authorization header
		tokenString := c.Request.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		if tokenString == "" {
			log.Warn().Msg("Authorization token is missing")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization token is required"})
			c.Abort()
			return
		}

		// Parse the JWT token
		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			log.Warn().Err(err).Msg("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Warn().Msg("Token is not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token is not valid"})
			c.Abort()
			return
		}

		// Extract userID from claims
		userID, ok := (*claims)["user_id"].(string)
		if !ok {
			log.Warn().Msg("Invalid token claims: user_id not found")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			c.Abort()
			return
		}

		// Fetch user by ID from the repository
		user, err := userRepo.FindByID(userID)
		if err != nil {
			log.Error().Str("user_id", userID).Err(err).Msg("User not found")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
			c.Abort()
			return
		}

		// Check if the user's restaurantID matches the requested restaurantID
		if user.RestaurantID.String() != restaurantIDStr {
			log.Info().Str("user_id", userID).Msg("Checking for admin permissions")
			isAdmin := false
			for _, perm := range user.Permissions {
				if perm.Name == "admin" {
					isAdmin = true
					break
				}
			}

			if !isAdmin {
				log.Warn().Str("user_id", userID).Msg("Unauthorized access attempt: user does not have admin permissions")
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access"})
				c.Abort()
				return
			}
		}

		// Set user and permissions in the context
		c.Set("user", user)
		c.Set("user_id", userID)
		c.Set("permissions", user.Permissions)

		log.Info().
			Str("user_id", userID).
			Int("num_permissions", len(user.Permissions)).
			Msg("User authenticated, restaurant and permissions set")

		c.Next()
	}
}
