package service

import (
	"context"
	"mime/multipart"
	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/data/response"

	"github.com/google/uuid"
)

type DishesService interface {
	Create(dish request.CreateDishRequest, userId uuid.UUID, requestId string, restaurantId uuid.UUID) (response.DishResponse, error)
	Update(dish request.UpdateDishRequest, userId uuid.UUID, requestId string, restaurantId string) (response.DishResponse, error)
	Delete(dishId uuid.UUID, restaurantId string, userId uuid.UUID, requestId string) error
	FindById(dishId uuid.UUID, restaurantId string, userId uuid.UUID, requestId string) (response.DishResponse, error)
	FindAll(page int, limit int, restaurantId string, userId uuid.UUID, requestId string) (response.DishListResponse, error)

	RateDish(dish request.RateDishRequest, userId uuid.UUID, dishId uuid.UUID, requestId string, restaurantId string) (response.RatingResponse, error)
	Search(searchTerm string, page int, limit int, restaurantId string, userId uuid.UUID, requestId string) (response.DishListResponse, error)
	UploadImageToS3(file multipart.FileHeader, ctx context.Context) (string, error)
}
