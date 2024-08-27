package repository

import (
	"fmt"
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DishesRepositoryImpl implements DishesRepository interface.
type DishesRepositoryImpl struct {
	Db *gorm.DB
}

// NewDishesRepositoryImpl creates a new instance of DishesRepositoryImpl.
func NewDishesRepositoryImpl(db *gorm.DB) DishesRepository {
	return &DishesRepositoryImpl{Db: db}
}

// Delete removes a dish from the database.
func (repo *DishesRepositoryImpl) Delete(dishId uuid.UUID, restaurantId string) error {
	var dish model.Dish
	result := repo.Db.Where("id = ? AND restaurant_id = ? AND deleted_at IS NULL", dishId, restaurantId).Delete(&dish)
	if result.Error != nil {
		log.Error().
			Str("dish_id", dishId.String()).
			Str("restaurant_id", restaurantId).
			Err(result.Error).
			Msg("Error deleting dish")
		return result.Error
	}
	log.Info().
		Str("dish_id", dishId.String()).
		Str("restaurant_id", restaurantId).
		Msg("Dish deleted successfully")
	return nil
}

// FindAll retrieves all dishes with pagination that are not soft-deleted for a specific restaurant.
func (repo *DishesRepositoryImpl) FindAll(page, limit int, restaurantId string) ([]model.Dish, int, error) {
	var dishes []model.Dish
	var total int64

	offset := (page - 1) * limit

	// Fetch total count of dishes for the specified restaurant
	countResult := repo.Db.Model(&model.Dish{}).
		Where("restaurant_id = ? AND deleted_at IS NULL", restaurantId).
		Count(&total)
	if countResult.Error != nil {
		log.Error().
			Str("restaurant_id", restaurantId).
			Err(countResult.Error).
			Msg("Error counting total dishes")
		return nil, 0, fmt.Errorf("error counting dishes: %w", countResult.Error)
	}

	// Fetch dishes with pagination for the specified restaurant
	result := repo.Db.Where("restaurant_id = ? AND deleted_at IS NULL", restaurantId).
		Limit(limit).
		Offset(offset).
		Find(&dishes)

	if result.Error != nil {
		log.Error().
			Str("restaurant_id", restaurantId).
			Err(result.Error).
			Msg("Error finding dishes")
		return nil, 0, fmt.Errorf("error finding dishes: %w", result.Error)
	}

	log.Info().
		Str("restaurant_id", restaurantId).
		Msgf("Retrieved dishes for page %d with limit %d successfully", page, limit)
	return dishes, int(total), nil
}

// FindById retrieves a dish by its ID for a specific restaurant and ensures it is not soft-deleted.
func (repo *DishesRepositoryImpl) FindById(dishId uuid.UUID, restaurantId string) (model.Dish, error) {
	var dish model.Dish
	result := repo.Db.Where("id = ? AND restaurant_id = ? AND deleted_at IS NULL", dishId, restaurantId).First(&dish)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Warn().
				Str("dish_id", dishId.String()).
				Str("restaurant_id", restaurantId).
				Msg("Dish not found")
			return model.Dish{}, fmt.Errorf("dish with ID %s not found for restaurant %s", dishId, restaurantId)
		}
		log.Error().
			Str("dish_id", dishId.String()).
			Str("restaurant_id", restaurantId).
			Err(result.Error).
			Msg("Error finding dish")
		return model.Dish{}, fmt.Errorf("error finding dish: %w", result.Error)
	}
	log.Info().
		Str("dish_id", dishId.String()).
		Str("restaurant_id", restaurantId).
		Msg("Dish retrieved successfully")
	return dish, nil
}

// Create adds a new dish to the database.
func (repo *DishesRepositoryImpl) Create(dish model.Dish, userId uuid.UUID) (model.Dish, error) {
	dish.ID = uuid.New()
	dish.CreatedById = userId
	result := repo.Db.Create(&dish)
	if result.Error != nil {
		log.Error().
			Str("dish_id", dish.ID.String()).
			Err(result.Error).
			Msg("Error creating dish")
		return model.Dish{}, result.Error
	}
	return dish, nil
}

// Update modifies an existing dish in the database.
func (repo *DishesRepositoryImpl) Update(dish model.Dish, userId uuid.UUID, restaurantId string) (model.Dish, error) {
	updateFields := map[string]interface{}{
		"LastUpdatedByID": userId,
	}

	if dish.Name != "" {
		updateFields["Name"] = dish.Name
	}
	if dish.Description != "" {
		updateFields["Description"] = dish.Description
	}
	if dish.Price != 0 {
		updateFields["Price"] = dish.Price
	}
	if dish.Image != "" {
		updateFields["Image"] = dish.Image
	}

	// Update the dish if it exists for the specified restaurant
	result := repo.Db.Model(&model.Dish{}).Where("id = ? AND restaurant_id = ?", dish.ID, restaurantId).Updates(updateFields)
	if result.Error != nil {
		log.Error().
			Str("dish_id", dish.ID.String()).
			Str("restaurant_id", restaurantId).
			Err(result.Error).
			Msg("Error updating dish")
		return model.Dish{}, fmt.Errorf("error updating dish: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Warn().
			Str("dish_id", dish.ID.String()).
			Str("restaurant_id", restaurantId).
			Msg("No rows affected, update might have failed")
		return model.Dish{}, fmt.Errorf("no rows affected, update might have failed")
	}

	var updatedDish model.Dish
	if err := repo.Db.Where("id = ? AND restaurant_id = ?", dish.ID, restaurantId).First(&updatedDish).Error; err != nil {
		log.Error().
			Str("dish_id", dish.ID.String()).
			Str("restaurant_id", restaurantId).
			Err(err).
			Msg("Error retrieving updated dish")
		return model.Dish{}, fmt.Errorf("error retrieving updated dish: %w", err)
	}

	log.Info().
		Str("dish_id", dish.ID.String()).
		Str("restaurant_id", restaurantId).
		Msg("Dish updated successfully")
	return updatedDish, nil
}

// RateDish adds a new rating to the database.
func (repo *DishesRepositoryImpl) RateDish(rating model.Rating) (model.Rating, error) {
	rating.ID = uuid.New()
	result := repo.Db.Create(&rating)
	if result.Error != nil {
		log.Error().
			Str("rating_id", rating.ID.String()).
			Err(result.Error).
			Msg("Error creating rating")
		return model.Rating{}, result.Error
	}
	log.Info().
		Str("rating_id", rating.ID.String()).
		Msg("Rating created successfully")
	return rating, nil
}

// Search finds dishes based on a search term in the dish name with pagination, filtering by restaurant ID.
func (repo *DishesRepositoryImpl) Search(searchTerm string, page, limit int, restaurantId string) ([]model.Dish, int, error) {
	var dishes []model.Dish
	var total int64

	// Prepare the search query
	query := "%" + searchTerm + "%"

	// Calculate the offset for pagination
	offset := (page - 1) * limit

	// Fetch the total count of matching dishes for pagination
	countResult := repo.Db.Model(&model.Dish{}).
		Where("deleted_at IS NULL AND name ILIKE ? AND restaurant_id = ?", query, restaurantId).
		Count(&total)
	if countResult.Error != nil {
		log.Error().
			Str("search_term", searchTerm).
			Str("restaurant_id", restaurantId).
			Err(countResult.Error).
			Msg("Error counting total dishes for search")
		return nil, 0, fmt.Errorf("error counting dishes: %w", countResult.Error)
	}

	// Fetch the paginated list of dishes
	result := repo.Db.Where("deleted_at IS NULL AND name ILIKE ? AND restaurant_id = ?", query, restaurantId).
		Limit(limit).
		Offset(offset).
		Find(&dishes)
	if result.Error != nil {
		log.Error().
			Str("search_term", searchTerm).
			Str("restaurant_id", restaurantId).
			Err(result.Error).
			Msg("Error searching dishes with pagination")
		return nil, 0, fmt.Errorf("error searching dishes: %w", result.Error)
	}

	log.Info().
		Str("search_term", searchTerm).
		Str("restaurant_id", restaurantId).
		Int("num_dishes_found", len(dishes)).
		Msg("Dishes searched successfully with pagination")

	return dishes, int(total), nil
}
