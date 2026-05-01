package model

import (
	"time"
)

type Credit struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	AccountID      string    `db:"account_id"`
	Principal      float64   `db:"principal"`
	InterestRate   float64   `db:"interest_rate"`
	TermMonths     int       `db:"term_months"`
	MonthlyPayment float64   `db:"monthly_payment"`
	RemainingDebt  float64   `db:"remaining_debt"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
}
