package dto

type AnalyticsResponse struct {
	Income           float64 `json:"income"`
	Expense          float64 `json:"expense"`
	CreditLoad       float64 `json:"credit_load"`
	PredictedBalance float64 `json:"predicted_balance"`
}
