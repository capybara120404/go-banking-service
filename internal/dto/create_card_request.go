package dto

type CreateCardRequest struct {
	AccountID string `json:"account_id" validate:"required,uuid"`
}

func (r *CreateCardRequest) Validate() error {
	if r.AccountID == "" {
		return NewErrorResponse("Account ID is required")
	}

	return nil
}
