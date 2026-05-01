package dto

type PredictBalanceResponse struct {
	AccountID        string  `json:"account_id"`
	CurrentBalance   float64 `json:"current_balance"`
	PredictedBalance float64 `json:"predicted_balance"`
	Days             int     `json:"days"`
}
