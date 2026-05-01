package model

import (
	"time"
)

type Transaction struct {
	ID          string    `db:"id"`
	SenderID    string    `db:"sender_id"`
	ReceiverID  string    `db:"receiver_id"`
	Amount      float64   `db:"amount"`
	Type        string    `db:"type"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}
