package response

import (
	"github.com/google/uuid"
)

// RestaurantResponse represents a single restaurant data in response.
type RestaurantResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	ImageUrl    string    `json:"imageUrl"`
}

// RestaurantListResponse represents a response containing a list of restaurants.
type RestaurantListResponse struct {
	Restaurants []RestaurantResponse `json:"restaurants"`  // List of restaurants for the current page
	CurrentPage int                  `json:"current_page"` // Current page number
	TotalPages  int                  `json:"total_pages"`  // Total number of pages
	TotalItems  int                  `json:"total_items"`  // Total number of items
}

// SearchRestaurantsResponse represents a response containing search results for restaurants.
type SearchRestaurantsResponse struct {
	Restaurants []RestaurantResponse `json:"restaurants"`
}
