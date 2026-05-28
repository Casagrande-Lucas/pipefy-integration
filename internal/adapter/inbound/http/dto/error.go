package dto

// ErrorResponse is the standard JSON error envelope returned on all failures.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse builds an ErrorResponse from an HTTP status code and message.
func NewErrorResponse(code int, message string) ErrorResponse {
	return ErrorResponse{Code: code, Message: message}
}
