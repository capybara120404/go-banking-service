package model

import (
	"time"
)

type Account struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Balance   float64   `db:"balance"`
	Currency  string    `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
}
