package helper

import (
	"strings"

	"go-fiber/app/model"
)

// ValidateLoginRequest memvalidasi login request
func ValidateLoginRequest(req *model.LoginRequest) []model.ValidationError {
	errors := []model.ValidationError{}

	// Validasi username
	if strings.TrimSpace(req.Username) == "" {
		errors = append(errors, model.ValidationError{
			Field:   "username",
			Message: "Username is required",
		})
	}

	// Validasi password
	if strings.TrimSpace(req.Password) == "" {
		errors = append(errors, model.ValidationError{
			Field:   "password",
			Message: "Password is required",
		})
	} else if len(req.Password) < 6 {
		errors = append(errors, model.ValidationError{
			Field:   "password",
			Message: "Password must be at least 6 characters",
		})
	}

	return errors
}
