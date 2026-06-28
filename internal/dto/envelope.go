package dto

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func Success(message string, data any) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

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
