package dto

import "time"

type CreditResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	AccountID      string    `json:"account_id"`
	Principal      float64   `json:"principal"`
	InterestRate   float64   `json:"interest_rate"`
	TermMonths     int       `json:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment"`
	RemainingDebt  float64   `json:"remaining_debt"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
