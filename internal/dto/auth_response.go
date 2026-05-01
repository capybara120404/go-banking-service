package dto

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}
