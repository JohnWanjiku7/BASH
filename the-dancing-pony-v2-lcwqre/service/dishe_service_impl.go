package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	utils "the-dancing-pony-v2-lcwqre/Utils"
	cache "the-dancing-pony-v2-lcwqre/caching"
	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/data/response"
	"the-dancing-pony-v2-lcwqre/model"
	"the-dancing-pony-v2-lcwqre/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// DishesServiceImpl provides the implementation for dish-related operations.
type DishesServiceImpl struct {
	DishesRepository repository.DishesRepository
	Validate         *validator.Validate
	S3Uploader       *utils.S3Uploader
}

// NewDishesServiceImpl creates a new instance of DishesServiceImpl.
func NewDishesServiceImpl(dishesRepository repository.DishesRepository, validate *validator.Validate, uploader *utils.S3Uploader) DishesService {
	return &DishesServiceImpl{
		DishesRepository: dishesRepository,
		Validate:         validate,
		S3Uploader:       uploader,
	}
}

// UploadImageToS3 uploads an image to S3 storage.
func (s *DishesServiceImpl) UploadImageToS3(fileHeader multipart.FileHeader, ctx context.Context) (string, error) {
	location, err := s.S3Uploader.UploadImage(fileHeader, ctx)
	if err != nil {
		return "", err
	}
	return location, nil
}

// Create adds a new dish to the repository and returns the created dish.
func (t *DishesServiceImpl) Create(dishRequest request.CreateDishRequest, userId uuid.UUID, requestID string, restaurantId uuid.UUID) (response.DishResponse, error) {
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Starting request validation for CreateDishRequest")

	dishModel := model.Dish{
		Name:         dishRequest.Name,
		Description:  dishRequest.Description,
		Price:        dishRequest.Price,
		Image:        dishRequest.ImageUrl,
		RestaurantID: restaurantId,
	}

	createdDish, err := t.DishesRepository.Create(dishModel, userId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Err(err).
			Msg("Error creating dish")
		return response.DishResponse{}, err
	}

	dishResponse := toDishResponse(createdDish)
	cacheKey := fmt.Sprintf("all_dishes_page_*_limit_*_restaurant_%s", restaurantId.String())
	cache.RedisClient.Del(context.Background(), cacheKey)
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Dish created successfully")
	return dishResponse, nil
}

// Delete removes a dish by ID.
func (t *DishesServiceImpl) Delete(dishId uuid.UUID, restaurantId string, userId uuid.UUID, requestID string) error {
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Starting dish deletion")

	if err := t.DishesRepository.Delete(dishId, restaurantId); err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Str("dish_id", dishId.String()).
			Err(err).
			Msg("Error deleting dish")
		return err
	}
	cacheKey := fmt.Sprintf("dish_%s_restaurant_%s", dishId.String(), restaurantId)
	cache.RedisClient.Del(context.Background(), cacheKey)
	cache.RedisClient.Del(context.Background(), fmt.Sprintf("all_dishes_page_*_limit_*_restaurant_%s", restaurantId))

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Str("dish_id", dishId.String()).
		Msg("Dish deleted successfully")
	return nil
}

// FindAll retrieves dishes with pagination, using cache if available.
func (t *DishesServiceImpl) FindAll(page, limit int, restaurantId string, userId uuid.UUID, requestID string) (response.DishListResponse, error) {
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Retrieving dishes with pagination")

	cacheKey := fmt.Sprintf("all_dishes_page_%d_limit_%d_restaurant_%s", page, limit, restaurantId)

	cachedData, err := cache.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var dishResponses []response.DishResponse
		if err := json.Unmarshal([]byte(cachedData), &dishResponses); err != nil {
			log.Error().
				Str("request_id", requestID).
				Str("user_id", userId.String()).
				Err(err).
				Msg("Error unmarshalling cached dishes")
			return response.DishListResponse{}, err
		}
		log.Info().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Msg("Dishes retrieved from cache")
		return response.DishListResponse{Dishes: dishResponses}, nil
	}

	dishes, total, err := t.DishesRepository.FindAll(page, limit, restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Err(err).
			Msg("Error retrieving dishes from repository")
		return response.DishListResponse{}, err
	}

	var dishResponses []response.DishResponse
	for _, dish := range dishes {
		dishResponses = append(dishResponses, toDishResponse(dish))
	}

	dishesJSON, err := json.Marshal(dishResponses)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Err(err).
			Msg("Error marshalling dishes for cache")
		return response.DishListResponse{}, err
	}

	cache.RedisClient.Set(context.Background(), cacheKey, dishesJSON, 10*time.Minute)
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Dishes cached successfully")

	totalPages := calculateTotalPages(total, limit)

	return response.DishListResponse{
		Dishes:      dishResponses,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  total,
	}, nil
}

// FindById retrieves a dish by its ID.
func (s *DishesServiceImpl) FindById(dishId uuid.UUID, restaurantId string, userId uuid.UUID, requestID string) (response.DishResponse, error) {
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Retrieving dish by ID")

	cacheKey := fmt.Sprintf("dish_%s_restaurant_%s", dishId.String(), restaurantId)

	cachedData, err := cache.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var dishResponse response.DishResponse
		if err := json.Unmarshal([]byte(cachedData), &dishResponse); err == nil {
			log.Info().
				Str("request_id", requestID).
				Str("user_id", userId.String()).
				Msg("Dish retrieved from cache")
			return dishResponse, nil
		}
	}

	dish, err := s.FindDishById(dishId, restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Err(err).
			Msg("Error finding dish by ID")
		return response.DishResponse{}, err
	}

	dishResponse := toDishResponse(dish)
	dishJSON, _ := json.Marshal(dishResponse)
	cache.RedisClient.Set(context.Background(), cacheKey, dishJSON, 10*time.Minute)

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Dish retrieved and cached successfully")
	return dishResponse, nil
}

// Update modifies an existing dish.
func (s *DishesServiceImpl) Update(dishUpdateRequest request.UpdateDishRequest, userId uuid.UUID, requestID string, restaurantId string) (response.DishResponse, error) {
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Starting dish update")

	dish, err := s.FindDishById(dishUpdateRequest.ID, restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Str("dish_id", dishUpdateRequest.ID.String()).
			Err(err).
			Msg("Error finding dish for update")
		return response.DishResponse{}, err
	}

	updatedDish, err := s.DishesRepository.Update(dish, userId, restaurantId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Str("dish_id", dishUpdateRequest.ID.String()).
			Err(err).
			Msg("Error updating dish")
		return response.DishResponse{}, fmt.Errorf("error updating dish: %w", err)
	}

	cacheKey := fmt.Sprintf("dish_%s_restaurant_%s", dishUpdateRequest.ID.String(), restaurantId)
	cache.RedisClient.Del(context.Background(), cacheKey)
	cache.RedisClient.Del(context.Background(), fmt.Sprintf("all_dishes_page_*_limit_*_restaurant_%s", restaurantId))

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Str("dish_id", dishUpdateRequest.ID.String()).
		Msg("Dish updated successfully")
	return toDishResponse(updatedDish), nil
}

// RateDish adds a rating to a dish.
func (s *DishesServiceImpl) RateDish(ratingRequest request.RateDishRequest, userId uuid.UUID, dishId uuid.UUID, requestID string, restaurantId string) (response.RatingResponse, error) {
	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Starting dish rating")

	dishRating := model.Rating{
		UserID: userId,
		Rating: ratingRequest.Rating,
		DishID: dishId,
	}

	rating, err := s.DishesRepository.RateDish(dishRating)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("user_id", userId.String()).
			Str("dish_id", dishId.String()).
			Err(err).
			Msg("Error rating dish")
		return response.RatingResponse{}, err
	}

	cache.RedisClient.Del(context.Background(), fmt.Sprintf("dish_%s_restaurant_%s", dishId.String(), restaurantId))
	cache.RedisClient.Del(context.Background(), fmt.Sprintf("all_dishes_page_*_limit_*_restaurant_%s", restaurantId))

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId.String()).
		Msg("Dish rated successfully")
	return response.RatingResponse{
		ID:     rating.ID,
		Rating: rating.Rating,
		DishId: rating.ID,
	}, nil
}

// FindDishById retrieves a dish by its ID from the repository.
func (s *DishesServiceImpl) FindDishById(dishId uuid.UUID, restaurantId string) (model.Dish, error) {
	dish, err := s.DishesRepository.FindById(dishId, restaurantId)
	if err != nil {
		log.Error().
			Str("dish_id", dishId.String()).
			Err(err).
			Msg("Error finding dish")
		return model.Dish{}, fmt.Errorf("error finding dish: %w", err)
	}
	return dish, nil
}

// Search searches for dishes matching the search term with pagination and caching.
func (s *DishesServiceImpl) Search(searchTerm string, page, limit int, restaurantId string, userId uuid.UUID, requestId string) (response.DishListResponse, error) {
	// Create a cache key based on search term and pagination parameters
	cacheKey := fmt.Sprintf("search_%s_page_%d_limit_%d,restaurant_%s", searchTerm, page, limit, restaurantId)

	// Attempt to retrieve cached data
	cachedData, err := cache.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var cachedResponse response.DishListResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err != nil {
			log.Error().
				Err(err).
				Msg("Error unmarshalling cached search results")
			return response.DishListResponse{}, err
		}
		log.Info().
			Msg("Search results retrieved from cache")
		return cachedResponse, nil
	}

	// Fetch paginated search results from the repository
	dishes, total, err := s.DishesRepository.Search(searchTerm, page, limit, restaurantId)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error retrieving search results from repository")
		return response.DishListResponse{}, err
	}

	// Convert model.Dish to response.DishResponse
	var dishResponses []response.DishResponse
	for _, dish := range dishes {
		dishResponses = append(dishResponses, toDishResponse(dish))
	}

	// Construct response with pagination details
	searchResponse := response.DishListResponse{
		Dishes:      dishResponses,
		CurrentPage: page,
		TotalPages:  calculateTotalPages(total, limit), // Calculate total pages based on total items
		TotalItems:  total,
	}

	// Cache the results
	dishesJSON, err := json.Marshal(searchResponse)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error marshalling search results for cache")
		return response.DishListResponse{}, err
	}
	cache.RedisClient.Set(context.Background(), cacheKey, dishesJSON, 10*time.Minute) // Adjust cache duration as needed
	log.Info().
		Msg("Search results cached successfully")

	return searchResponse, nil
}

// calculateTotalPages calculates the total number of pages based on total items and items per page.
func calculateTotalPages(totalItems, limit int) int {
	if limit == 0 {
		return 0
	}
	return (totalItems + limit - 1) / limit // Ceiling division
}

// Helper function to convert model.Dish to response.DishResponse
func toDishResponse(dish model.Dish) response.DishResponse {
	return response.DishResponse{
		ID:          dish.ID,
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		ImageUrl:    dish.Image,
	}
}
