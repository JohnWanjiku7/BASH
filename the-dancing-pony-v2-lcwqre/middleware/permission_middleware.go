package middleware

import (
	"net/http"
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// PermissionMiddleware2 checks if the user has one of the required permissions
func PermissionMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve permissions from the context
		permissions, exists := c.Get("permissions")
		if !exists {
			log.Info().Msg("No permissions found in context. User might not be authenticated or permissions not set.")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: no permissions found"})
			c.Abort()
			return
		}

		// Assert the type to []model.Permission
		permArray, ok := permissions.([]model.Permission)
		if !ok {
			log.Error().Msgf("Permissions retrieved but not in the expected format: %T. Expected []model.Permission. Ensure that permissions are properly set in the context.", permissions)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: unexpected permissions format"})
			c.Abort()
			return
		}

		// Extract and log permission names
		var permNames []string
		for _, perm := range permArray {
			permNames = append(permNames, perm.Name)
		}

		if len(permNames) > 0 {
			log.Info().Strs("permissions", permNames).Msg("Permissions info")
		} else {
			log.Warn().Msg("Permissions array is empty. This might indicate that the user has no permissions assigned.")
		}

		// Create a map for quick lookup of required permissions
		requiredPermissionsMap := make(map[string]bool)
		for _, perm := range requiredPermissions {
			requiredPermissionsMap[perm] = true
		}

		// Check if user has any of the required permissions
		hasPermission := false
		for _, perm := range permArray {
			if requiredPermissionsMap[perm.Name] {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			log.Warn().
				Strs("required_permissions", requiredPermissions).
				Strs("user_permissions", permNames).
				Msg("User does not have any of the required permissions. Access denied.")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		// If the user has any of the required permissions, continue to the next handler
		c.Next()
	}
}
