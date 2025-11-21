package model

// APIResponse struktur standar untuk semua API response
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ValidationError untuk error validasi
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse helper untuk response sukses
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Status: "success",
		Data:   data,
	}
}

// ErrorResponse helper untuk response error
func ErrorResponse(message string, err interface{}) APIResponse {
	return APIResponse{
		Status:  "error",
		Message: message,
		Error:   err,
	}
}
