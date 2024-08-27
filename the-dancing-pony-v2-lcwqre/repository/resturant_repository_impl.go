package repository

import (
	"fmt"
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// RestaurantsRepositoryImpl implements RestaurantsRepository interface.
type RestaurantsRepositoryImpl struct {
	Db *gorm.DB
}

// NewRestaurantsRepositoryImpl creates a new instance of RestaurantsRepositoryImpl.
func NewRestaurantsRepositoryImpl(db *gorm.DB) RestaurantsRepository {
	return &RestaurantsRepositoryImpl{Db: db}
}

// Delete removes a restaurant from the database.
func (repo *RestaurantsRepositoryImpl) Delete(restaurantId uuid.UUID) error {
	var restaurant model.Restaurant
	result := repo.Db.Where("id = ? AND deleted_at IS NULL", restaurantId).Delete(&restaurant)
	if result.Error != nil {
		log.Error().
			Str("restaurant_id", restaurantId.String()).
			Err(result.Error).
			Msg("Error deleting restaurant")
		return result.Error
	}
	log.Info().
		Str("restaurant_id", restaurantId.String()).
		Msg("Restaurant deleted successfully")
	return nil
}

// FindAll retrieves all restaurants with pagination that are not soft-deleted.
func (repo *RestaurantsRepositoryImpl) FindAll(page, limit int) ([]model.Restaurant, int, error) {
	var restaurants []model.Restaurant
	var total int64

	offset := (page - 1) * limit

	// Fetch total count of restaurants
	countResult := repo.Db.Model(&model.Restaurant{}).Where("deleted_at IS NULL").Count(&total)
	if countResult.Error != nil {
		log.Error().
			Err(countResult.Error).
			Msg("Error counting total restaurants")
		return nil, 0, fmt.Errorf("error counting restaurants: %w", countResult.Error)
	}

	// Fetch restaurants with pagination
	result := repo.Db.Where("deleted_at IS NULL").
		Limit(limit).
		Offset(offset).
		Find(&restaurants)

	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Msg("Error finding restaurants")
		return nil, 0, fmt.Errorf("error finding restaurants: %w", result.Error)
	}

	log.Info().
		Msgf("Retrieved restaurants for page %d with limit %d successfully", page, limit)
	return restaurants, int(total), nil
}

// FindById retrieves a restaurant by its ID.
func (repo *RestaurantsRepositoryImpl) FindById(restaurantId uuid.UUID) (model.Restaurant, error) {
	var restaurant model.Restaurant
	result := repo.Db.Where("id = ? AND deleted_at IS NULL", restaurantId).First(&restaurant)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Warn().
				Str("restaurant_id", restaurantId.String()).
				Msg("Restaurant not found")
			return model.Restaurant{}, fmt.Errorf("restaurant with ID %s not found", restaurantId)
		}
		log.Error().
			Str("restaurant_id", restaurantId.String()).
			Err(result.Error).
			Msg("Error finding restaurant")
		return model.Restaurant{}, fmt.Errorf("error finding restaurant: %w", result.Error)
	}
	log.Info().
		Str("restaurant_id", restaurantId.String()).
		Msg("Restaurant retrieved successfully")
	return restaurant, nil
}

// Create adds a new restaurant to the database.
func (repo *RestaurantsRepositoryImpl) Create(restaurant model.Restaurant) (model.Restaurant, error) {
	restaurant.ID = uuid.New()
	result := repo.Db.Create(&restaurant)
	if result.Error != nil {
		log.Error().
			Str("restaurant_id", restaurant.ID.String()).
			Err(result.Error).
			Msg("Error creating restaurant")
		return model.Restaurant{}, result.Error
	}
	log.Info().
		Str("restaurant_id", restaurant.ID.String()).
		Msg("Restaurant created successfully")
	return restaurant, nil
}

// Update modifies an existing restaurant in the database.
func (repo *RestaurantsRepositoryImpl) Update(restaurant model.Restaurant) (model.Restaurant, error) {
	updateFields := make(map[string]interface{})

	if restaurant.Name != "" {
		updateFields["Name"] = restaurant.Name
	}
	if restaurant.Description != "" {
		updateFields["Description"] = restaurant.Description
	}
	if restaurant.Location != "" {
		updateFields["Location"] = restaurant.Location
	}
	if restaurant.ImageUrl != "" {
		updateFields["ImageUrl"] = restaurant.ImageUrl
	}

	result := repo.Db.Model(&model.Restaurant{}).Where("id = ?", restaurant.ID).Updates(updateFields)
	if result.Error != nil {
		log.Error().
			Str("restaurant_id", restaurant.ID.String()).
			Err(result.Error).
			Msg("Error updating restaurant")
		return model.Restaurant{}, fmt.Errorf("error updating restaurant: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Warn().
			Str("restaurant_id", restaurant.ID.String()).
			Msg("No rows affected, update might have failed")
		return model.Restaurant{}, fmt.Errorf("no rows affected, update might have failed")
	}

	var updatedRestaurant model.Restaurant
	if err := repo.Db.First(&updatedRestaurant, restaurant.ID).Error; err != nil {
		log.Error().
			Str("restaurant_id", restaurant.ID.String()).
			Err(err).
			Msg("Error retrieving updated restaurant")
		return model.Restaurant{}, fmt.Errorf("error retrieving updated restaurant: %w", err)
	}

	log.Info().
		Str("restaurant_id", restaurant.ID.String()).
		Msg("Restaurant updated successfully")
	return updatedRestaurant, nil
}

// Search finds restaurants based on a search term in the restaurant name with pagination.
func (repo *RestaurantsRepositoryImpl) Search(searchTerm string, page, limit int) ([]model.Restaurant, int, error) {
	var restaurants []model.Restaurant
	var total int64

	// Prepare the search query
	query := "%" + searchTerm + "%"

	// Calculate the offset for pagination
	offset := (page - 1) * limit

	// Fetch the total count of matching restaurants for pagination
	countResult := repo.Db.Model(&model.Restaurant{}).
		Where("deleted_at IS NULL AND name ILIKE ?", query).
		Count(&total)
	if countResult.Error != nil {
		log.Error().
			Err(countResult.Error).
			Msg("Error counting total restaurants for search")
		return nil, 0, fmt.Errorf("error counting restaurants: %w", countResult.Error)
	}

	// Fetch the paginated list of restaurants
	result := repo.Db.Where("deleted_at IS NULL AND name ILIKE ?", query).
		Limit(limit).
		Offset(offset).
		Find(&restaurants)
	if result.Error != nil {
		log.Error().
			Err(result.Error).
			Msg("Error searching restaurants with pagination")
		return nil, 0, fmt.Errorf("error searching restaurants: %w", result.Error)
	}

	log.Info().
		Str("search_term", searchTerm).
		Int("num_restaurants_found", len(restaurants)).
		Msg("Restaurants searched successfully with pagination")

	return restaurants, int(total), nil
}
