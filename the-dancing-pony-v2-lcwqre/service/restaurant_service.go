package service

import (
	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/data/response"

	"github.com/google/uuid"
)

// RestaurantsService defines the interface for restaurant-related business logic operations.
type RestaurantsService interface {
	// Create adds a new restaurant and returns the created restaurant.
	Create(restaurantRequest request.CreateRestaurantRequest, requestID string) (response.RestaurantResponse, error)

	// Update modifies an existing restaurant and returns the updated restaurant.
	Update(restaurantUpdateRequest request.UpdateRestaurantRequest) (response.RestaurantResponse, error)

	// Delete removes a restaurant by its ID.
	Delete(restaurantId uuid.UUID) error

	// FindAll retrieves a list of all restaurants with pagination.
	FindAll(page, limit int) (response.RestaurantListResponse, error)

	// FindById retrieves a specific restaurant by its ID.
	FindById(restaurantId uuid.UUID) (response.RestaurantResponse, error)

	// Search searches for restaurants based on a search term with pagination.
	Search(searchTerm string, page, limit int) (response.RestaurantListResponse, error)
}
