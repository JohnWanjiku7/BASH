package request

type RegisterRequest struct {
	Name        string   `json:"name" validate:"required"`
	Email       string   `json:"email" validate:"required,email"`
	Password    string   `json:"password" validate:"required"`
	Permissions []string `json:"permissions" validate:"required"`
}

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
