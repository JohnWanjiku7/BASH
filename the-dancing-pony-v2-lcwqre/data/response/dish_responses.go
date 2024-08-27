package response

import "github.com/google/uuid"

// Common response structure for success and error messages.
type APIResponse struct {
	Status  string      `json:"status"` // "success" or "error"
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse represents a successful response with data.
func SuccessResponse(status string, message string, data interface{}) APIResponse {
	return APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse represents an error response with a message.
func ErrorResponse(message string) APIResponse {
	return APIResponse{
		Status:  "error",
		Message: message,
	}
}

// DishResponse represents a single dish data in response.
type DishResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageUrl    string    `json:"imageUrl"`
}

// DishListResponse represents a response containing a list of dishes.
type DishListResponse struct {
	Dishes      []DishResponse `json:"dishes"`       // List of dishes for the current page
	CurrentPage int            `json:"current_page"` // Current page number
	TotalPages  int            `json:"total_pages"`  // Total number of pages
	TotalItems  int            `json:"total_items"`  // Total number of items
}

// SearchDishesResponse represents a response containing search results.
type SearchDishesResponse struct {
	Dishes []DishResponse `json:"dishes"`
}

// RatingResponse represents a response for a rating action.
type RatingResponse struct {
	ID     uuid.UUID `json:"id"`
	Rating int       `json:"rating"`
	DishId uuid.UUID `json:"dishId"`
}

/*// SuccessResponse represents a successful response with data.
func SuccessResponse(status string, message string, data interface{}) APIResponse {
	return APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse represents an error response with a message.
func ErrorResponse(message string) APIResponse {
	return APIResponse{
		Status:  "error",
		Message: message,
	}
} */
