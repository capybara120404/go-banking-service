package dto

import "strings"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return NewErrorResponse("Email is required")
	} else if strings.Contains(r.Email, "@") == false {
		return NewErrorResponse("Email must be valid")
	}
	if r.Password == "" {
		return NewErrorResponse("Password is required")
	} else if len(r.Password) < 8 {
		return NewErrorResponse("Password must be at least 8 characters long")
	}

	return nil
}
