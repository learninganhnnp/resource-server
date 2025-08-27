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

