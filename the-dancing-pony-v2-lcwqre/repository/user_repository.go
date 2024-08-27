package repository

import (
	"the-dancing-pony-v2-lcwqre/model"

	"github.com/google/uuid"
)

// UserRepository defines the methods for interacting with the User model.
type UserRepository interface {
	Create(user model.User) error
	FindByEmail(email string, restaurantId uuid.UUID) (*model.User, error)
	FindByID(id string) (model.User, error)
	FindPermission(name string) (model.Permission, error)
}
