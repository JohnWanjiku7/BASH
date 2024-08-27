package response

// AuthResponse represents the response for successful authentication.
type AuthResponse struct {
	Token string `json:"token"`
}
