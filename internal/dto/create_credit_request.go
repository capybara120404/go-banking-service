package dto

type CreateCreditRequest struct {
	AccountID  string  `json:"account_id" validate:"required,uuid"`
	Principal  float64 `json:"principal" validate:"required,gt=0"`
	TermMonths int     `json:"term_months" validate:"required,gt=0,lte=360"`
}

func (r *CreateCreditRequest) Validate() error {
	if r.AccountID == "" {
		return NewErrorResponse("Account ID is required")
	}
	if r.Principal <= 0 {
		return NewErrorResponse("Principal must be greater than zero")
	}
	if r.TermMonths <= 0 {
		return NewErrorResponse("Term must be greater than zero months")
	}

	return nil
}
