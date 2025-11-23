package service

import (
    "fmt"

    "go-fiber/app/model"
    "go-fiber/app/repository"
    "go-fiber/helper"
)

type AuthService struct {
    userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
    return &AuthService{userRepo: userRepo}
}

// LOGIN
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {

    // Cari user
    user, err := s.userRepo.FindByUsernameOrEmail(req.Username)
    if err != nil {
        return nil, fmt.Errorf("invalid username or password")
    }

    // Cek password
    if !helper.CheckPasswordHash(req.Password, user.PasswordHash) {
        return nil, fmt.Errorf("invalid username or password")
    }

    // Cek aktif
    if !user.IsActive {
        return nil, fmt.Errorf("user not active")
    }

    // Ambil permissions
    perms, _ := s.userRepo.GetUserPermissions(user.RoleID)

    // Generate JWT
    access, _  := helper.GenerateAccessToken(user, perms)
    refresh, _ := helper.GenerateRefreshToken(user.ID)

    return &model.LoginResponse{
        Token:        access,
        RefreshToken: refresh,
        User: model.UserResponse{
            ID:          user.ID,
            Username:    user.Username,
            FullName:    user.FullName,
            Role:        user.RoleName,
            Permissions: perms,
        },
    }, nil
}

// REFRESH TOKEN
func (s *AuthService) RefreshToken(refreshToken string) (*model.LoginResponse, error) {

    userID, err := helper.ValidateRefreshToken(refreshToken)
    if err != nil {
        return nil, err
    }

    // Ambil user
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, err
    }

    if !user.IsActive {
        return nil, fmt.Errorf("user inactive")
    }

    perms, _ := s.userRepo.GetUserPermissions(user.RoleID)

    // Generate new tokens
    access, _  := helper.GenerateAccessToken(user, perms)
    newRefresh, _ := helper.GenerateRefreshToken(user.ID)

    return &model.LoginResponse{
        Token:        access,
        RefreshToken: newRefresh,
        User: model.UserResponse{
            ID:          user.ID,
            Username:    user.Username,
            FullName:    user.FullName,
            Role:        user.RoleName,
            Permissions: perms,
        },
    }, nil
}

// LOGOUT
func (s *AuthService) Logout() error {
    // Karena refresh token tersimpan di cookie
    // logout hanya akan menghapus cookie di route
    return nil
}

// GetUserProfile mengambil data profil berdasarkan userID
func (s *AuthService) GetUserProfile(userID string) (*model.UserResponse, error) {

	// Ambil user dari repository
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Ambil permission berdasarkan role user
	perms, err := s.userRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions")
	}

	// Bentuk response
	return &model.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: perms,
	}, nil
}
