package dto

import "time"

type PaymentScheduleResponse struct {
	ID             string    `json:"id"`
	CreditID       string    `json:"credit_id"`
	PaymentDate    time.Time `json:"payment_date"`
	Amount         float64   `json:"amount"`
	IsPaid         bool      `json:"is_paid"`
	LateFeeApplied bool      `json:"late_fee_applied"`
}
