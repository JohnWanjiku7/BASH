package repository

import (
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/google/uuid"
)

type DishesRepository interface {
	Create(dish model.Dish, userId uuid.UUID) (model model.Dish, err error)
	Update(dish model.Dish, userId uuid.UUID, restaurantId string) (model model.Dish, err error)
	Delete(dishId uuid.UUID, restaurantId string) (err error)
	FindById(dish_id uuid.UUID, restaurantId string) (dish model.Dish, err error)
	FindAll(page, limit int, restaurantId string) (returnDishes []model.Dish, count int, err error)

	RateDish(dish model.Rating) (model model.Rating, err error)
	Search(searchTerm string, page int, limit int, restaurantId string) ([]model.Dish, int, error)
}
