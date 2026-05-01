package dto

type CreateAccountRequest struct {
	Currency string `json:"currency" validate:"required,eq=RUB"`
}

func (r *CreateAccountRequest) Validate() error {
	if r.Currency != "RUB" {
		return NewErrorResponse("Currency must be RUB")
	}

	return nil
}
