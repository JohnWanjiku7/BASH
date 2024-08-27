package repository

import (
	"errors"
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UserRepositoryImpl implements UserRepository interface.
type UserRepositoryImpl struct {
	Db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepositoryImpl.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{Db: db}
}

// Create inserts a new user into the database.
func (repo *UserRepositoryImpl) Create(user model.User) error {
	user.ID = uuid.New()
	result := repo.Db.Create(&user)
	if result.Error != nil {
		log.Error().
			Str("user_id", user.ID.String()).
			Err(result.Error).
			Msg("Error creating user")
		return result.Error
	}
	log.Info().
		Str("user_id", user.ID.String()).
		Msg("User created successfully")
	return nil
}

// FindByEmail retrieves a user by their email and ensures the user belongs to the specified restaurant.
func (repo *UserRepositoryImpl) FindByEmail(email string, restaurantId uuid.UUID) (*model.User, error) {
	var user model.User
	// Query for a user with the specified email and restaurant ID

	result := repo.Db.Where("email = ? AND restaurant_id = ?", email, restaurantId).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Record not found, which is expected when the user does not exist
			log.Info().
				Str("email", email).
				Str("restaurant_id", restaurantId.String()).
				Msg("User not found")
			return nil, errors.New("User not found")
		}

		// Log the error and return it
		log.Error().
			Str("email", email).
			Str("restaurant_id", restaurantId.String()).
			Err(result.Error).
			Msg("Error finding user by email and restaurant ID")
		return nil, result.Error
	}

	// Return a pointer to the found user
	log.Info().
		Str("email", email).
		Str("restaurant_id", restaurantId.String()).
		Msg("User found successfully")
	return &user, nil
}

// FindByID retrieves a user by their ID along with their permissions.
func (repo *UserRepositoryImpl) FindByID(id string) (model.User, error) {
	var user model.User
	result := repo.Db.Preload("Permissions").Where("id = ? AND deleted_at IS NULL", id).First(&user)
	if result.Error != nil {
		log.Error().
			Str("id", id).
			Err(result.Error).
			Msg("Error finding user by ID")
		return user, result.Error
	}
	return user, nil
}

// FindPermission retrieves a permission by its name.
func (repo *UserRepositoryImpl) FindPermission(name string) (model.Permission, error) {
	var permission model.Permission
	result := repo.Db.Where("name = ?", name).First(&permission)
	if result.Error != nil {
		log.Error().
			Str("permission_name", name).
			Err(result.Error).
			Msg("Error finding permission")
		return permission, result.Error
	}
	return permission, nil
}
