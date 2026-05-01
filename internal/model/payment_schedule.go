package model

import (
	"time"
)

type PaymentSchedule struct {
	ID             string    `db:"id"`
	CreditID       string    `db:"credit_id"`
	PaymentDate    time.Time `db:"payment_date"`
	Amount         float64   `db:"amount"`
	IsPaid         bool      `db:"is_paid"`
	LateFeeApplied bool      `db:"late_fee_applied"`
}
