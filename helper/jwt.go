package helper

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-fiber/app/model"
)

// GenerateAccessToken membuat JWT access token
func GenerateAccessToken(user *model.User, permissions []string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	expiresIn := os.Getenv("JWT_EXPIRES_IN")

	// Parse duration (default 24 jam)
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		duration = 24 * time.Hour
	}

	// Buat claims
	claims := jwt.MapClaims{
		"userId":      user.ID,
		"username":    user.Username,
		"role":        user.RoleName,
		"permissions": permissions,
		"type":        "access",
		"exp":         time.Now().Add(duration).Unix(),
		"iat":         time.Now().Unix(),
	}

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken membuat JWT refresh token
func GenerateRefreshToken(userID string) (string, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	expiresIn := os.Getenv("JWT_REFRESH_EXPIRES_IN")

	// Parse duration (default 7 hari)
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		duration = 168 * time.Hour // 7 hari
	}

	// Buat claims
	claims := jwt.MapClaims{
		"userId": userID,
		"type":   "refresh",
		"exp":    time.Now().Add(duration).Unix(),
		"iat":    time.Now().Unix(),
	}

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateAccessToken memvalidasi access token
func ValidateAccessToken(tokenString string) (*model.JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validasi signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validasi tipe token
	if claims["type"] != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Convert permissions
	permissions := []string{}
	if perms, ok := claims["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if perm, ok := p.(string); ok {
				permissions = append(permissions, perm)
			}
		}
	}

	return &model.JWTClaims{
		UserID:      claims["userId"].(string),
		Username:    claims["username"].(string),
		Role:        claims["role"].(string),
		Permissions: permissions,
		Type:        claims["type"].(string),
	}, nil
}

// ValidateRefreshToken memvalidasi refresh token
func ValidateRefreshToken(tokenString string) (string, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Validasi tipe token
	if claims["type"] != "refresh" {
		return "", fmt.Errorf("invalid token type")
	}

	return claims["userId"].(string), nil
}