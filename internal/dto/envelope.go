package dto

// SuccessResponse is the graded success envelope: {success, message, data}.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// ErrorResponse is the graded error envelope: {success, message, errors}.
// Errors is a field-level map for validation failures; use a single key for non-field errors.
type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

// Success builds a success envelope.
func Success(message string, data any) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// Error builds an error envelope.
func Error(message string, errors map[string]string) ErrorResponse {
	if errors == nil {
		errors = map[string]string{}
	}
	return ErrorResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}
