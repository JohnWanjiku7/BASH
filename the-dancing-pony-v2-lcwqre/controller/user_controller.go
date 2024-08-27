package controller

import (
	"net/http"
	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/data/response"
	"the-dancing-pony-v2-lcwqre/helper"
	"the-dancing-pony-v2-lcwqre/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// AuthController handles authentication-related HTTP requests.
type AuthController struct {
	AuthService service.AuthService
	Validate    *validator.Validate
}

// NewAuthController creates a new instance of AuthController.
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		AuthService: authService,
		Validate:    validator.New(),
	}
}

// Register handles user registration.
func (ctrl *AuthController) Register(ctx *gin.Context) {
	requestID := ctx.GetString("request_id")

	log.Info().
		Str("request_id", requestID).
		Msg("Registration request processing started")

	var req request.RegisterRequest

	// Validate the registration request
	if !helper.ValidateRequest(ctx, &req, ctrl.Validate, requestID) {
		log.Warn().
			Str("request_id", requestID).
			Msg("Registration request validation failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration data"})
		return
	}

	log.Info().
		Str("request_id", requestID).
		Msg("Registration payload successfully validated")

	resturantId, err := helper.GetRestaurant(ctx)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Failed to get resturant")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get resturant"})
		return
	}
	// Perform registration
	err = ctrl.AuthService.Register(req, resturantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msgf("Registration request error: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

// Login handles user login and returns a JWT token.
func (ctrl *AuthController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	requestID := ctx.GetString("request_id")

	log.Info().
		Str("request_id", requestID).
		Msg("Login request processing started")

	// Validate the login request
	if !helper.ValidateRequest(ctx, &req, ctrl.Validate, requestID) {
		log.Warn().
			Str("request_id", requestID).
			Msg("Login request validation failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}
	resturantId, err := helper.GetRestaurant(ctx)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Failed to get resturant")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get resturant"})
		return
	}

	// Perform login and generate token
	token, err := ctrl.AuthService.Login(req, resturantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msgf("Login request error: %s", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Login failed"})
		return
	}

	ctx.JSON(http.StatusOK, response.AuthResponse{Token: token})
}
