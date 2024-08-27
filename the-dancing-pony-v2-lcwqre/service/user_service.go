package service

import (
	"the-dancing-pony-v2-lcwqre/data/request"

	"github.com/google/uuid"
)

// AuthService defines the methods for user authentication.
type AuthService interface {
	Register(req request.RegisterRequest, resturantId uuid.UUID) error
	Login(req request.LoginRequest, resturantId uuid.UUID) (string, error)
}
