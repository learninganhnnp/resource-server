package dto

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents error details in API responses
type APIError struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// SuccessResponseWithMessage creates a successful API response with message
func SuccessResponseWithMessage(data interface{}, message string) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(code string, message string, details string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error: &APIError{
			Code:    code,
			Details: details,
		},
	}
}