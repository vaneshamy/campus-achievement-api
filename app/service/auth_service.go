package service

import (
	"fmt"
	"time"

	"go-fiber/app/model"
	"go-fiber/app/repository"
	"go-fiber/helper"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// Login melakukan proses autentikasi user
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// 1. Cari user berdasarkan username atau email
	var user *model.User
	var err error

	// Cek apakah input adalah email atau username
	if contains(req.Username, "@") {
		user, err = s.userRepo.FindByEmail(req.Username)
	} else {
		user, err = s.userRepo.FindByUsername(req.Username)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 2. Validasi password
	if !helper.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 3. Cek status aktif user
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// 4. Load permissions user
	permissions, err := s.userRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions")
	}

	// 5. Generate access token
	accessToken, err := helper.GenerateAccessToken(user, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	// 6. Generate refresh token
	refreshToken, err := helper.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// 7. Simpan refresh token ke database
	expiresIn := helper.ParseDuration("JWT_REFRESH_EXPIRES_IN", 168*time.Hour)
	token := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(expiresIn),
	}

	if err := s.tokenRepo.SaveRefreshToken(token); err != nil {
		return nil, fmt.Errorf("failed to save refresh token")
	}

	// 8. Return response
	response := &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: model.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Role:        user.RoleName,
			Permissions: permissions,
		},
	}

	return response, nil
}

// RefreshToken melakukan refresh access token
func (s *AuthService) RefreshToken(refreshTokenStr string) (*model.LoginResponse, error) {
	// 1. Validasi refresh token
	userID, err := helper.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// 2. Cek apakah refresh token ada di database
	storedToken, err := s.tokenRepo.FindRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refresh token")
	}
	if storedToken == nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// 3. Load user data
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 4. Cek status aktif user
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// 5. Load permissions
	permissions, err := s.userRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions")
	}

	// 6. Generate access token baru
	accessToken, err := helper.GenerateAccessToken(user, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	// 7. Return response (refresh token tetap sama)
	response := &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshTokenStr,
		User: model.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Role:        user.RoleName,
			Permissions: permissions,
		},
	}

	return response, nil
}

// Logout menghapus refresh token dari database
func (s *AuthService) Logout(refreshToken string) error {
	return s.tokenRepo.DeleteRefreshToken(refreshToken)
}

// GetUserProfile mendapatkan profil user dari token
func (s *AuthService) GetUserProfile(userID string) (*model.UserResponse, error) {
	// 1. Load user data
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 2. Load permissions
	permissions, err := s.userRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions")
	}

	// 3. Return response
	response := &model.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: permissions,
	}

	return response, nil
}

// Helper function
func contains(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && 
		   (str == substr || (len(str) > len(substr) && 
		   (str[:len(substr)] == substr || str[len(str)-len(substr):] == substr || 
		   containsSubstr(str, substr))))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}