package model

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Dish struct {
	gorm.Model
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Price           float64    `json:"price"`
	Image           string     `json:"image"`
	CreatedById     uuid.UUID  `json:"created_by_id"`                            // Foreign key for the user who created the dish
	CreatedBy       User       `gorm:"foreignKey:CreatedById;references:ID"`     // Belongs to User
	LastUpdatedByID *uuid.UUID `json:"last_updated_by_id"`                       // Foreign key for the user who last updated the dish
	LastUpdatedBy   User       `gorm:"foreignKey:LastUpdatedByID;references:ID"` // Belongs to User
	Ratings         []Rating   `gorm:"foreignKey:DishID"`                        // One-to-many relationship with ratings
	RestaurantID    uuid.UUID  `gorm:"index;not null"`                           // Foreign key for the restaurant
	Restaurant      Restaurant `gorm:"foreignKey:RestaurantID"`
}

// Rating model
type Rating struct {
	gorm.Model
	DishID uuid.UUID `json:"dish_id"`
	Dish   Dish      `gorm:"foreignKey:DishID;references:ID"` // Belongs to Dish
	UserID uuid.UUID `json:"user_id"`
	User   User      `gorm:"foreignKey:UserID;references:ID"` // Belongs to User
	Rating int       `json:"rating"`
}

// User represents a user entity in the system.
type User struct {
	gorm.Model
	Name         string       `gorm:"not null" json:"name"`
	Email        string       `gorm:"not null,uniqueIndex" json:"email"`
	Password     string       `gorm:"not null" json:"-"`
	Permissions  []Permission `gorm:"many2many:user_permissions"`
	Dishes       []Dish       `gorm:"foreignKey:CreatedById"` // One-to-many relationship with dishes created by the user
	Ratings      []Rating     `gorm:"foreignKey:UserID"`      // One-to-many relationship with ratings given by the user
	RestaurantID uuid.UUID    `gorm:"index"`                  // Foreign key for the restaurant
	Restaurant   Restaurant   `gorm:"foreignKey:RestaurantID"`
}

// Permission represents a permission assigned to a user.
type Permission struct {
	gorm.Model
	Name string `json:"name" validate:"required,unique"`
}

// Restaurant represents a restaurant entity.
type Restaurant struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `json:"description" validate:"required,min=1,max=200"`
	Location    string `json:"location" validate:"required,min=1,max=200"`
	ImageUrl    string `json:"imageUrl" validate:"required,min=1,max=200"`
}

// SetPassword hashes the password for the User
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword compares a given password with the hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
