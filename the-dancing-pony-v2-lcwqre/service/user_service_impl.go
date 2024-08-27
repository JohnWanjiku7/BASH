package service

import (
	"errors"
	"time"

	"the-dancing-pony-v2-lcwqre/data/request"
	"the-dancing-pony-v2-lcwqre/helper"
	"the-dancing-pony-v2-lcwqre/model"
	"the-dancing-pony-v2-lcwqre/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// AuthServiceImpl implements AuthService interface.
type AuthServiceImpl struct {
	UserRepo repository.UserRepository
}

// NewAuthService creates a new instance of AuthServiceImpl.
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &AuthServiceImpl{UserRepo: userRepo}
}

// Register registers a new user.
func (s *AuthServiceImpl) Register(req request.RegisterRequest, restaurantId uuid.UUID) error {
	log.Info().Str("email", req.Email).Msg("Processing registration request")

	// Check if user already exists
	existingUser, err := s.UserRepo.FindByEmail(req.Email, restaurantId)
	if err != nil {
		existingUser = nil

	}
	if existingUser != nil {
		log.Error().Str("email", req.Email).Msg("User with that email already exists")
		return errors.New("user with that email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Str("email", req.Email).Err(err).Msg("Failed to hash password")
		return err
	}

	// Fetch permissions
	var permissions []model.Permission
	for _, permName := range req.Permissions {
		perm, err := s.UserRepo.FindPermission(permName)
		if err != nil {
			log.Error().Str("permission", permName).Err(err).Msg("Failed to find permission")
			return err
		}
		permissions = append(permissions, perm)
	}

	// Create user object
	user := model.User{
		Name:         req.Name,
		Email:        req.Email,
		Password:     string(hashedPassword),
		Permissions:  permissions,
		RestaurantID: restaurantId,
	}

	// Create new user
	if err := s.UserRepo.Create(user); err != nil {
		log.Error().Str("email", req.Email).Err(err).Msg("Failed to create user")
		return err
	}

	log.Info().Str("email", req.Email).Msg("User registered successfully")
	return nil
}

// Login logs in a user and returns a JWT token.
func (s *AuthServiceImpl) Login(req request.LoginRequest, restaurantId uuid.UUID) (string, error) {
	user, err := s.UserRepo.FindByEmail(req.Email, restaurantId)
	if err != nil {
		log.Error().Str("email", req.Email).Err(err).Msg("User not found")
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Error().Str("email", req.Email).Msg("Invalid credentials provided")
		return "", errors.New("invalid email or password")
	}

	token, err := generateToken(user.ID)
	if err != nil {
		log.Error().Str("email", req.Email).Err(err).Msg("Failed to generate token")
		return "", err
	}

	log.Info().Str("email", req.Email).Msg("User logged in successfully")
	return token, nil
}

// generateToken generates a JWT token for the given user ID.
func generateToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(helper.GetJWTSecret()))
}
