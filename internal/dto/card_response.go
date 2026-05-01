package dto

import "time"

type CardResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	AccountID      string    `json:"account_id"`
	NumberLastFour string    `json:"number_last_four"`
	ExpiryDate     string    `json:"expiry_date"`
	CreatedAt      time.Time `json:"created_at"`
}
