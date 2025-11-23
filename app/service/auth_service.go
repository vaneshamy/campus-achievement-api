package service

import (
	"fmt"

	"go-fiber/app/model"
	"go-fiber/app/repository"
	"go-fiber/helper"
)

// AuthService tidak lagi membutuhkan TokenRepository (no DB for refresh tokens)
type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Login memverifikasi user, mengembalikan access token dan refresh token
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// Cari user by username, jika tidak ada coba by email
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		// coba email
		user, err = s.userRepo.FindByEmail(req.Username)
		if err != nil {
			return nil, fmt.Errorf("invalid credentials")
		}
	}

	// Cek active
	if !user.IsActive {
		return nil, fmt.Errorf("user is not active")
	}

	// Cek password (asumsi helper.CheckPasswordHash)
	if ok := helper.CheckPasswordHash(req.Password, user.PasswordHash); !ok {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Ambil permissions user
	perms, err := s.userRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions: %v", err)
	}

	// Generate access token
	accessToken, err := helper.GenerateAccessToken(user, perms)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (JWT) — disimpan di cookie pada route
	refreshToken, err := helper.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Siapkan response user shape
	userResp := model.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: perms,
	}

	return &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken, // route akan menaruhnya di cookie, jangan kirim lewat JSON kalau tak ingin
		User:         userResp,
	}, nil
}

// RefreshToken menerima refresh token string (dari cookie), memvalidasi dan create access token baru.
// Opsi: melakukan rotation refresh token (mengembalikan refresh token baru)
func (s *AuthService) RefreshToken(refreshToken string) (*model.LoginResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token missing")
	}

	userID, err := helper.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Ambil user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user not active")
	}

	// Ambil permissions
	perms, err := s.userRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, err
	}

	// Buat access token baru
	accessToken, err := helper.GenerateAccessToken(user, perms)
	if err != nil {
		return nil, err
	}

	// Opsi: rotate refresh token untuk keamanan (lebih baik)
	newRefreshToken, err := helper.GenerateRefreshToken(user.ID)
	if err != nil {
		// jika gagal generate refresh baru, tetap beri access token lama
		newRefreshToken = refreshToken // fallback, meski idealnya kita handle differently
	}

	userResp := model.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: perms,
	}

	return &model.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		User:         userResp,
	}, nil
}

// Logout tidak perlu menghapus DB token — cukup clear cookie di route
func (s *AuthService) Logout(refreshToken string) error {
	// Karena tidak ada penyimpanan server-side, logout server hanya no-op
	// (Cookie akan dihapus oleh route)
	return nil
}

// GetUserProfile mengambil profil user (dipanggil dari route /auth/profile)
func (s *AuthService) GetUserProfile(userID string) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}
