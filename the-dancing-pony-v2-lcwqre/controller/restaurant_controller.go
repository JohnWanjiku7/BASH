package controller

import (
	"net/http"
	"strconv"
	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/data/response"
	"the-dancing-pony-v2-lcwqre/helper"
	"the-dancing-pony-v2-lcwqre/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type RestaurantsController struct {
	RestaurantsService service.RestaurantsService
	Validate           *validator.Validate
}

func NewRestaurantsController(service service.RestaurantsService) *RestaurantsController {
	return &RestaurantsController{
		RestaurantsService: service,
		Validate:           validator.New(),
	}
}

// Create handles the creation of a restaurant.
func (controller *RestaurantsController) Create(ctx *gin.Context) {
	requestID := ctx.GetString("request_id")
	log.Info().
		Str("request_id", requestID).
		Msg("Create restaurant request received")

	var createRestaurantRequest request.CreateRestaurantRequest

	if !helper.ValidateRequest(ctx, &createRestaurantRequest, controller.Validate, requestID) {
		log.Error().
			Str("request_id", requestID).
			Msg("Request validation failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return
	}

	createdRestaurant, err := controller.RestaurantsService.Create(createRestaurantRequest, requestID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Error creating restaurant")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating restaurant"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Restaurant created successfully",
		Status:  "Ok",
		Data:    createdRestaurant,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
	log.Info().
		Str("request_id", requestID).
		Str("restaurant_id", createdRestaurant.ID.String()).
		Msg("Restaurant created successfully")
}

// Update updates a restaurant.
func (controller *RestaurantsController) Update(ctx *gin.Context) {
	requestID := ctx.GetString("request_id")
	log.Info().
		Str("request_id", requestID).
		Msg("Update restaurant request received")

	var updateRestaurantRequest request.UpdateRestaurantRequest
	if !helper.ValidateRequest(ctx, &updateRestaurantRequest, controller.Validate, requestID) {
		log.Warn().
			Str("request_id", requestID).
			Msg("Request validation failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return
	}

	restaurantId := ctx.Param("restaurantId")
	id, err := uuid.Parse(restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("restaurant_id", restaurantId).
			Msg("Invalid restaurant ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID format"})
		return
	}

	updateRestaurantRequest.ID = id
	updatedRestaurant, err := controller.RestaurantsService.Update(updateRestaurantRequest)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("restaurant_id", restaurantId).
			Msg("Error updating restaurant")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating restaurant"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Restaurant updated successfully",
		Status:  "Ok",
		Data:    updatedRestaurant,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// Delete removes a restaurant.
func (controller *RestaurantsController) Delete(ctx *gin.Context) {
	requestID := ctx.GetString("request_id")
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Delete Restaurant request")

	restaurantId := ctx.Param("restaurantId")
	id, err := uuid.Parse(restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("restaurant_id", restaurantId).
			Msg("Invalid restaurant ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID format"})
		return
	}

	err = controller.RestaurantsService.Delete(id)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("restaurant_id", restaurantId).
			Msg("Error deleting restaurant")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting restaurant"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Restaurant deleted successfully",
		Status:  "Ok",
		Data:    nil,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// FindById retrieves a restaurant by its ID.
func (controller *RestaurantsController) FindById(ctx *gin.Context) {
	requestID := ctx.GetString("request_id")
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Find Restaurant by ID request")

	restaurantId := ctx.Param("restaurantId")
	id, err := uuid.Parse(restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("restaurant_id", restaurantId).
			Msg("Invalid restaurant ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID format"})
		return
	}

	restaurant, err := controller.RestaurantsService.FindById(id)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("restaurant_id", restaurantId).
			Msg("Restaurant not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Restaurant found",
		Status:  "Ok",
		Data:    restaurant,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// FindAll retrieves all restaurants with pagination.
func (controller *RestaurantsController) FindAll(ctx *gin.Context) {
	requestID := ctx.GetString("request_id")
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Find All Restaurants request")

	// Extract pagination parameters
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Fetch restaurants with pagination
	restaurantListResponse, err := controller.RestaurantsService.FindAll(page, limit)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Error retrieving restaurants")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving restaurants"})
		return
	}

	// Construct paginated response
	webResponse := response.APIResponse{
		Message: "Restaurants retrieved successfully",
		Status:  "Ok",
		Data:    restaurantListResponse,
	}

	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}
