package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cache "the-dancing-pony-v2-lcwqre/caching"
	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/data/response"
	"the-dancing-pony-v2-lcwqre/model"
	"the-dancing-pony-v2-lcwqre/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RestaurantsServiceImpl provides the implementation for restaurant-related operations.
type RestaurantsServiceImpl struct {
	RestaurantsRepository repository.RestaurantsRepository
	Validate              *validator.Validate
}

// NewRestaurantsServiceImpl creates a new instance of RestaurantsServiceImpl.
func NewRestaurantsServiceImpl(restaurantsRepository repository.RestaurantsRepository, validate *validator.Validate) RestaurantsService {
	return &RestaurantsServiceImpl{
		RestaurantsRepository: restaurantsRepository,
		Validate:              validate,
	}
}

// Create adds a new restaurant to the repository and returns the created restaurant.
func (s *RestaurantsServiceImpl) Create(restaurantRequest request.CreateRestaurantRequest, requestID string) (response.RestaurantResponse, error) {
	log.Info().
		Str("request_id", requestID).
		Msg("Starting request validation for CreateRestaurantRequest")

	if err := s.Validate.Struct(restaurantRequest); err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Validation failed for CreateRestaurantRequest")
		return response.RestaurantResponse{}, err
	}

	restaurantModel := model.Restaurant{
		Name:        restaurantRequest.Name,
		Description: restaurantRequest.Description,
		Location:    restaurantRequest.Location,
		ImageUrl:    restaurantRequest.ImageUrl,
	}

	createdRestaurant, err := s.RestaurantsRepository.Create(restaurantModel)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Error creating restaurant")
		return response.RestaurantResponse{}, err
	}

	restaurantResponse := toRestaurantResponse(createdRestaurant)
	log.Info().
		Str("request_id", requestID).
		Msg("Restaurant created successfully")
	return restaurantResponse, nil
}

// Delete removes a restaurant by ID.
func (s *RestaurantsServiceImpl) Delete(restaurantId uuid.UUID) error {
	if err := s.RestaurantsRepository.Delete(restaurantId); err != nil {
		log.Error().
			Str("restaurant_id", restaurantId.String()).
			Err(err).
			Msg("Error deleting restaurant")
		return err
	}
	log.Info().
		Str("restaurant_id", restaurantId.String()).
		Msg("Restaurant deleted successfully")
	return nil
}

// FindAll retrieves restaurants with pagination, using cache if available.
func (s *RestaurantsServiceImpl) FindAll(page, limit int) (response.RestaurantListResponse, error) {
	cacheKey := fmt.Sprintf("all_restaurants_page_%d_limit_%d", page, limit)

	// Attempt to retrieve cached data
	cachedData, err := cache.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var restaurantResponses []response.RestaurantResponse
		if err := json.Unmarshal([]byte(cachedData), &restaurantResponses); err != nil {
			log.Error().
				Err(err).
				Msg("Error unmarshalling cached restaurants")
			return response.RestaurantListResponse{}, err
		}
		log.Info().
			Msg("Restaurants retrieved from cache")
		return response.RestaurantListResponse{Restaurants: restaurantResponses}, nil
	}

	// Fetch restaurants from repository with pagination
	restaurants, total, err := s.RestaurantsRepository.FindAll(page, limit)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error retrieving restaurants from repository")
		return response.RestaurantListResponse{}, err
	}

	var restaurantResponses []response.RestaurantResponse
	for _, restaurant := range restaurants {
		restaurantResponses = append(restaurantResponses, toRestaurantResponse(restaurant))
	}

	// Cache the fetched restaurants
	restaurantsJSON, err := json.Marshal(restaurantResponses)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error marshalling restaurants for cache")
		return response.RestaurantListResponse{}, err
	}

	cache.RedisClient.Set(context.Background(), cacheKey, restaurantsJSON, 10*time.Minute) // Increased cache duration
	log.Info().
		Msg("Restaurants cached successfully")

	// Calculate total pages
	totalPages := (total + limit - 1) / limit // This calculates the ceiling of total/limit

	return response.RestaurantListResponse{
		Restaurants: restaurantResponses,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  total,
	}, nil
}

// FindById retrieves a restaurant by its ID.
func (s *RestaurantsServiceImpl) FindById(restaurantId uuid.UUID) (response.RestaurantResponse, error) {
	restaurant, err := s.FindRestaurantById(restaurantId)
	if err != nil {
		return response.RestaurantResponse{}, err
	}
	return toRestaurantResponse(restaurant), nil
}

// Update modifies an existing restaurant.
func (s *RestaurantsServiceImpl) Update(restaurantUpdateRequest request.UpdateRestaurantRequest) (response.RestaurantResponse, error) {
	restaurant, err := s.FindRestaurantById(restaurantUpdateRequest.ID)
	if err != nil {
		return response.RestaurantResponse{}, err
	}

	updatedRestaurant, err := s.RestaurantsRepository.Update(restaurant)
	if err != nil {
		log.Error().
			Str("restaurant_id", restaurantUpdateRequest.ID.String()).
			Err(err).
			Msg("Error updating restaurant")
		return response.RestaurantResponse{}, fmt.Errorf("error updating restaurant: %w", err)
	}

	return toRestaurantResponse(updatedRestaurant), nil
}

// FindRestaurantById retrieves a restaurant by its ID from the repository.
func (s *RestaurantsServiceImpl) FindRestaurantById(restaurantId uuid.UUID) (model.Restaurant, error) {
	restaurant, err := s.RestaurantsRepository.FindById(restaurantId)
	if err != nil {
		log.Error().
			Str("restaurant_id", restaurantId.String()).
			Err(err).
			Msg("Error finding restaurant")
		return model.Restaurant{}, fmt.Errorf("error finding restaurant: %w", err)
	}
	return restaurant, nil
}

// Search searches for restaurants matching the search term with pagination and caching.
func (s *RestaurantsServiceImpl) Search(searchTerm string, page, limit int) (response.RestaurantListResponse, error) {
	// Create a cache key based on search term and pagination parameters
	cacheKey := fmt.Sprintf("search_%s_page_%d_limit_%d", searchTerm, page, limit)

	// Attempt to retrieve cached data
	cachedData, err := cache.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var cachedResponse response.RestaurantListResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err != nil {
			log.Error().
				Err(err).
				Msg("Error unmarshalling cached search results")
			return response.RestaurantListResponse{}, err
		}
		log.Info().
			Msg("Search results retrieved from cache")
		return cachedResponse, nil
	}

	// Fetch paginated search results from the repository
	restaurants, total, err := s.RestaurantsRepository.Search(searchTerm, page, limit)
	if err != nil {
		return response.RestaurantListResponse{}, err
	}

	// Convert model.Restaurant to response.RestaurantResponse
	var restaurantResponses []response.RestaurantResponse
	for _, restaurant := range restaurants {
		restaurantResponses = append(restaurantResponses, toRestaurantResponse(restaurant))
	}

	// Construct response with pagination details
	searchResponse := response.RestaurantListResponse{
		Restaurants: restaurantResponses,
		CurrentPage: page,
		TotalPages:  calculateTotalPages(total, limit), // Calculate total pages based on total items
		TotalItems:  total,
	}

	// Cache the results
	restaurantsJSON, err := json.Marshal(searchResponse)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error marshalling search results for cache")
		return response.RestaurantListResponse{}, err
	}
	cache.RedisClient.Set(context.Background(), cacheKey, restaurantsJSON, 10*time.Minute) // Adjust cache duration as needed
	log.Info().
		Msg("Search results cached successfully")

	return searchResponse, nil
}

// Helper function to convert model.Restaurant to response.RestaurantResponse
func toRestaurantResponse(restaurant model.Restaurant) response.RestaurantResponse {
	return response.RestaurantResponse{
		ID:          restaurant.ID,
		Name:        restaurant.Name,
		Description: restaurant.Description,
		Location:    restaurant.Location,
		ImageUrl:    restaurant.ImageUrl,
	}
}
