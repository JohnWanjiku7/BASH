package request

import (
	"github.com/google/uuid"
)

// RestaurantBase defines common fields and validation for restaurant-related requests.
type RestaurantBase struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"required,min=1,max=200"`
	Location    string `json:"location" validate:"required,min=1,max=200"`
	ImageUrl    string `json:"imageUrl" validate:"required,min=1,max=200"`
}

// CreateRestaurantRequest represents a request to create a restaurant.
type CreateRestaurantRequest struct {
	RestaurantBase
}

// UpdateRestaurantRequest represents a request to update a restaurant.
type UpdateRestaurantRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
	RestaurantBase
}

// ListRestaurantRequestList represents a list of restaurant creation requests.
type ListRestaurantRequestList struct {
	Restaurants []CreateRestaurantRequest `json:"restaurants"`
}

// DeleteRestaurantRequest represents a request to delete a restaurant.
type DeleteRestaurantRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
}

// RateRestaurantRequest represents a request to rate a restaurant.
type RateRestaurantRequest struct {
	Rating int `json:"rating" validate:"required,gt=0,lt=6"`
}
