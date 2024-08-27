package request

import (
	"github.com/go-playground/validator/v10" // Import validator package for validation
	"github.com/google/uuid"
)

// DishBase defines common fields and validation for dish-related requests.
type DishBase struct {
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description string  `json:"description" validate:"required,min=1,max=200"`
	Price       float64 `json:"price" validate:"required,gt=0"` // Use 'gt=0' for price greater than 0
	ImageUrl    string  `json:"imageUrl" validate:"required,min=1,max=200"`
}

// CreateDishRequest represents a request to create a dish.
type CreateDishRequest struct {
	UserID uuid.UUID `json:"userId"`
	DishBase
}

type CreateDishAPIRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description string  `json:"description" validate:"required,min=1,max=200"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

// UpdateDishRequest represents a request to update a dish.
type UpdateDishRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
	DishBase
}

// ListDishRequestList represents a list of dish creation requests.
type ListDishRequestList struct {
	Dishes []CreateDishRequest `json:"dishes"`
}

// DeleteDishRequest represents a request to delete a dish.
type DeleteDishRequest struct {
	ID uuid.UUID `json:"id" validate:"required"`
}

// RateDishRequest represents a request to rate a dish.
type RateDishRequest struct {
	Rating int `json:"rating" validate:"required,gt=0,lt=6"`
}

// Validate validates the CreateDishRequestList fields.
func (c *ListDishRequestList) Validate() error {
	validate := validator.New()
	for _, dish := range c.Dishes {
		if err := validate.Struct(dish); err != nil {
			return err
		}
	}
	return nil
}
