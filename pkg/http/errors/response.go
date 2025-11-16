package errors

// SuccessResponse represents a successful HTTP response
type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

// ErrorResponse represents an error HTTP response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
