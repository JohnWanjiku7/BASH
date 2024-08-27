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

type DishesController struct {
	DishesService service.DishesService
	Validate      *validator.Validate
}

func NewDishesController(service service.DishesService) *DishesController {
	return &DishesController{
		DishesService: service,
		Validate:      validator.New(),
	}
}

// Create handles the creation of a dish.
func (controller *DishesController) Create(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}
	restaurantUUID, err := uuid.Parse(restaurantId)
	if err != nil {

		return
	}
	// Parse multipart form data with a max memory size of 32MB
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Error parsing form data: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Extract form fields
	name := ctx.Request.FormValue("name")
	description := ctx.Request.FormValue("description")
	price := ctx.Request.FormValue("price")
	if name == "" || description == "" || price == "" {
		log.Error().Str("request_id", requestID).Msg("Missing required form fields")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}
	// Parse price to integer
	priceInt, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Error().Str("request_id", requestID).Msg("Invalid price format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
		return
	}
	var createDishesRequest request.CreateDishAPIRequest

	createDishesRequest.Name = name
	createDishesRequest.Description = description
	createDishesRequest.Price = priceInt

	// Parse the multipart form to get the file and other fields
	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		log.Error().Str("request_id", requestID).Msg("Failed to get image file")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		return
	}
	defer file.Close()

	// Upload image to S3
	imageURL, err := controller.DishesService.UploadImageToS3(*header, ctx)
	if err != nil {
		log.Error().Str("request_id", requestID).Msg("Failed to upload image to S3")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}
	var dishRequest request.CreateDishRequest
	dishRequest.Name = createDishesRequest.Name
	dishRequest.Description = createDishesRequest.Description
	dishRequest.Price = createDishesRequest.Price
	dishRequest.ImageUrl = imageURL

	createdDish, err := controller.DishesService.Create(dishRequest, userId, requestID, restaurantUUID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Error creating dish")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating dish"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Dish created successfully",
		Status:  "Ok",
		Data:    createdDish,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
	log.Info().
		Str("request_id", requestID).
		Str("dish_id", createdDish.ID.String()).
		Msg("Dish created successfully")
}

// Update updates a dish.
func (controller *DishesController) Update(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}

	var updateDishesRequest request.UpdateDishRequest

	dishId := ctx.Param("dishId")
	id, err := uuid.Parse(dishId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Invalid dish ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dish ID format"})
		return
	}

	// Parse multipart form data with a max memory size of 32MB
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Error parsing form data: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Extract form fields
	name := ctx.Request.FormValue("name")
	description := ctx.Request.FormValue("description")
	price := ctx.Request.FormValue("price")
	if name == "" || description == "" || price == "" {
		log.Error().Str("request_id", requestID).Msg("Missing required form fields")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}
	// Parse price to integer
	priceInt, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Error().Str("request_id", requestID).Msg("Invalid price format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
		return
	}
	var updateDisheRequest request.UpdateDishRequest

	updateDisheRequest.Name = name
	updateDisheRequest.Description = description
	updateDisheRequest.Price = priceInt

	// Parse the multipart form to get the file and other fields
	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		log.Error().Str("request_id", requestID).Msg("Failed to get image file")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		return
	}
	defer file.Close()

	// Upload image to S3
	imageURL, err := controller.DishesService.UploadImageToS3(*header, ctx)
	if err != nil {
		log.Error().Str("request_id", requestID).Msg("Failed to upload image to S3")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}
	updateDisheRequest.ID = id
	updateDisheRequest.ImageUrl = imageURL

	updatedDish, err := controller.DishesService.Update(updateDishesRequest, userId, requestID, restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Error updating dish")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating dish"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Dish updated successfully",
		Status:  "Ok",
		Data:    updatedDish,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// Delete removes a dish.
func (controller *DishesController) Delete(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Delete Dish request")

	dishId := ctx.Param("dishId")
	id, err := uuid.Parse(dishId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Invalid dish ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dish ID format"})
		return
	}
	err = controller.DishesService.Delete(id, restaurantId, userId, requestID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Error deleting dish")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting dish"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Dish deleted successfully",
		Status:  "Ok",
		Data:    nil,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// FindById retrieves a dish by its ID.
func (controller *DishesController) FindById(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Find Dish by ID request")

	dishId := ctx.Param("dishId")
	id, err := uuid.Parse(dishId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Invalid dish ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dish ID format"})
		return
	}
	dish, err := controller.DishesService.FindById(id, restaurantId, userId, requestID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Dish not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Dish not found"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Dish found",
		Status:  "Ok",
		Data:    dish,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// FindAll retrieves all dishes with pagination.
func (controller *DishesController) FindAll(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Find All Dishes request")

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

	// Fetch dishes with pagination
	dishListResponse, err := controller.DishesService.FindAll(page, limit, restaurantId, userId, requestID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Error retrieving dishes")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving dishes"})
		return
	}

	// Construct paginated response
	webResponse := response.APIResponse{
		Message: "Dishes retrieved successfully",
		Status:  "Ok",
		Data:    dishListResponse,
	}

	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// RateDish allows a user to rate a dish.
func (controller *DishesController) RateDish(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}
	log.Info().
		Str("request_id", requestID).
		Msg("Processing Rate Dish request")

	var rateDishesRequest request.RateDishRequest

	if !helper.ValidateRequest(ctx, &rateDishesRequest, controller.Validate, requestID) {
		log.Warn().
			Str("request_id", requestID).
			Msg("Request validation failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return
	}

	dishId := ctx.Param("dishId")
	id, err := uuid.Parse(dishId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Invalid dish ID format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dish ID format"})
		return
	}

	dishRating, err := controller.DishesService.RateDish(rateDishesRequest, userId, id, requestID, restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("dish_id", dishId).
			Msg("Error rating dish")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error rating dish"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Dish successfully rated",
		Status:  "Ok",
		Data:    dishRating,
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// Search finds dishes based on a search term with pagination.
func (controller *DishesController) Search(ctx *gin.Context) {
	requestID, userId, restaurantId, err := helper.ExtractRequestData(ctx)
	if err != nil {
		return
	}
	searchTerm := ctx.Query("searchTerm")
	if searchTerm == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Search term is required"})
		return
	}

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

	log.Info().
		Str("request_id", requestID).
		Str("search_term", searchTerm).
		Int("page", page).
		Int("limit", limit).
		Msg("Searching for dishes")

	dishListResponse, err := controller.DishesService.Search(searchTerm, page, limit, restaurantId, userId, requestID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("search_term", searchTerm).
			Int("page", page).
			Int("limit", limit).
			Msg("Error searching dishes")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error searching dishes"})
		return
	}

	webResponse := response.APIResponse{
		Message: "Dishes retrieved successfully",
		Status:  "Ok",
		Data:    dishListResponse,
	}

	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, webResponse)
}

// Handles pagination extraction
func ExtractPagination(ctx *gin.Context) (int, int) {
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

	return page, limit
}
