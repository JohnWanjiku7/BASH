package repository

import (
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/google/uuid"
)

// RestaurantsRepository defines the interface for restaurant-related data operations.
type RestaurantsRepository interface {
	// Create adds a new restaurant to the database.
	Create(restaurant model.Restaurant) (model.Restaurant, error)

	// Update modifies an existing restaurant in the database.
	Update(restaurant model.Restaurant) (model.Restaurant, error)

	// Delete removes a restaurant from the database.
	Delete(restaurantId uuid.UUID) error

	// FindAll retrieves all restaurants with pagination.
	FindAll(page, limit int) ([]model.Restaurant, int, error)

	// FindById retrieves a restaurant by its ID.
	FindById(restaurantId uuid.UUID) (model.Restaurant, error)

	// Search finds restaurants based on a search term in the restaurant name with pagination.
	Search(searchTerm string, page, limit int) ([]model.Restaurant, int, error)
}
