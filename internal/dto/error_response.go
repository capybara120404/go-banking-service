package dto

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		Message: message,
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (r *ErrorResponse) Error() string {
	return r.Message
}
