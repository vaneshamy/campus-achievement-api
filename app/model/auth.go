package model

// LoginRequest DTO untuk login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse DTO untuk login response
type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

// RefreshTokenRequest DTO untuk refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// JWTClaims struktur untuk JWT payload
type JWTClaims struct {
	UserID      string   `json:"userId"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Type        string   `json:"type"` // "access" atau "refresh"
}

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

