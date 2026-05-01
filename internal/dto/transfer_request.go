package dto

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id" validate:"required,uuid"`
	ToAccountID   string  `json:"to_account_id" validate:"required,uuid"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
}

func (r *TransferRequest) Validate() error {
	if r.FromAccountID == "" {
		return NewErrorResponse("From Account ID is required")
	}
	if r.ToAccountID == "" {
		return NewErrorResponse("To Account ID is required")
	}
	if r.Amount <= 0 {
		return NewErrorResponse("Amount must be greater than zero")
	}

	return nil
}
