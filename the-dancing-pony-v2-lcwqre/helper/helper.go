package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Extracts requestID, userId, and restaurantId
func ExtractRequestData(ctx *gin.Context) (string, uuid.UUID, string, error) {
	requestID := ctx.GetString("request_id")

	userId, err := GetUser(ctx)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Failed to get user ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user ID"})
		return "", uuid.Nil, "", err
	}

	restaurantId, err := GetRestaurantAsString(ctx)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Failed to get restaurant")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get restaurant"})
		return "", uuid.Nil, "", err
	}

	return requestID, userId, restaurantId, nil
}

// HandleValidationError logs validation errors in a structured format.
func HandleValidationError(c *gin.Context, err error, requestID string) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var allErrors []string
		for _, e := range validationErrors {
			fieldName := e.Field()
			tag := e.Tag()

			var errorMessage string
			switch tag {
			case "url":
				errorMessage = fmt.Sprintf("Field '%s' must be a valid URL", fieldName)
			default:
				errorMessage = fmt.Sprintf("Field '%s' failed validation on '%s' tag", fieldName, tag)
			}

			allErrors = append(allErrors, errorMessage)
		}

		// Log structured validation errors
		response := map[string][]string{"errors": allErrors}
		jsonResponse, jsonErr := json.Marshal(response)
		if jsonErr != nil {
			log.Error().
				Str("request_id", requestID).
				Err(jsonErr).
				Msg("Error marshalling JSON for validation errors")
		} else {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("validation_errors", string(jsonResponse)).
				Msg("Validation errors encountered")
		}
	} else {
		// Log unexpected errors
		log.Error().
			Str("request_id", c.GetString("request_id")).
			Msgf("Unexpected error: %s - invalid payload", err.Error())
	}
}

// ValidateRequest validates the request structure and its contents.
func ValidateRequest(c *gin.Context, req interface{}, validate *validator.Validate, requestID string) bool {
	log.Info().
		Str("request_id", requestID).
		Msg("Starting request validation")

	if validate == nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Validator instance is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return false
	}

	// Bind the JSON payload to the request struct
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Error reading request body")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body)) // Reset the body after reading

	if err := c.ShouldBindJSON(req); err != nil {
		log.Error().
			Str("request_id", requestID).
			Msgf("Failed to bind request JSON: %s. Payload: %s", err.Error(), string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return false
	}

	// Validate the struct
	if err := validate.Struct(req); err != nil {
		HandleValidationError(c, err, requestID)
		log.Error().
			Str("request_id", requestID).
			Msgf("Validation failed: %s. Payload: %s", err.Error(), string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return false
	}

	log.Info().
		Str("request_id", requestID).
		Msg("Request validation successful")
	return true
}

// LogInformation logs an error with additional context and sends a response to the client.
func LogInformation(ctx *gin.Context, statusCode int, message string, err error, requestID string) {
	log.Error().
		Str("request_id", requestID).
		Err(err).
		Msg(message)
	ctx.JSON(statusCode, gin.H{"error": message})
}

// GetUser retrieves the user UUID from the Gin context.
func GetUser(ctx *gin.Context) (uuid.UUID, error) {
	// Retrieve the user ID from the Gin context
	userID, exists := ctx.Get("user_id")
	if !exists {
		return uuid.UUID{}, errors.New("user ID not found in context")
	}

	// Assert the type of userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		return uuid.UUID{}, errors.New("invalid user ID format")
	}

	// Parse the user ID string to UUID
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, errors.New("failed to parse user ID")
	}

	return userUUID, nil
}

// GetRestaurant retrieves the restaurant UUID from the Gin context.
func GetRestaurant(ctx *gin.Context) (uuid.UUID, error) {
	// Retrieve the restaurant ID from the Gin context
	restaurantID, exists := ctx.Get("restaurant")
	if !exists {
		return uuid.UUID{}, errors.New("restaurant ID not found in context")
	}

	// Assert the type of restaurantID to string
	restaurantIDStr, ok := restaurantID.(string)
	if !ok {
		return uuid.UUID{}, errors.New("invalid restaurant ID format")
	}

	// Parse the restaurant ID string to UUID
	restaurantUUID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		return uuid.UUID{}, errors.New("failed to parse restaurant ID")
	}

	return restaurantUUID, nil
}

// GetRestaurantAsString retrieves the restaurant ID from the Gin context as a string.
func GetRestaurantAsString(ctx *gin.Context) (string, error) {
	// Retrieve the restaurant ID from the Gin context
	restaurantID, exists := ctx.Get("restaurant")
	if !exists {
		return "", errors.New("restaurant ID not found in context")
	}

	// Assert the type of restaurantID to string
	restaurantIDStr, ok := restaurantID.(string)
	if !ok {
		return "", errors.New("invalid restaurant ID format")
	}

	// Return the restaurant ID as a string
	return restaurantIDStr, nil
}

// GetJWTSecret retrieves the JWT secret from the environment variable.
func GetJWTSecret() string {
	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		errMsg := "JWT_SECRET environment variable is not set"
		log.Error().Msg(errMsg)
		return ""
	}
	return secret
}
