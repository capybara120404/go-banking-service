package model

import (
	"time"
)

type Card struct {
	ID              string    `db:"id"`
	UserID          string    `db:"user_id"`
	AccountID       string    `db:"account_id"`
	NumberEncrypted []byte    `db:"number_encrypted"`
	ExpiryEncrypted []byte    `db:"expiry_encrypted"`
	CVVHash         []byte    `db:"cvv_hash"`
	HMAC            string    `db:"hmac"`
	CreatedAt       time.Time `db:"created_at"`
}
