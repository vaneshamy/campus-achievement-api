package service

import (
    "fmt"
    "time"

    "github.com/gofiber/fiber/v2"

    "go-fiber/app/model"
    "go-fiber/app/repository"
    "go-fiber/helper"
)

type AuthService struct {
    authRepo repository.AuthRepositoryInterface  
}

func NewAuthService(authRepo repository.AuthRepositoryInterface) *AuthService {
    return &AuthService{authRepo: authRepo}
}


// LOGIN
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {

	user, err := s.authRepo.FindByUsernameOrEmail(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !helper.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user not active")
	}

	perms, _ := s.authRepo.GetUserPermissions(user.RoleID)

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


func (s *AuthService) HandleLogin(c *fiber.Ctx) error {

    var req model.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.JSON(model.ErrorResponse("Invalid body", err))
    }

    if errors := helper.ValidateLoginRequest(&req); len(errors) > 0 {
        return c.JSON(model.ErrorResponse("Validation failed", errors))
    }

    // --- logic login ---
    res, err := s.Login(&req)
    if err != nil {
        return c.JSON(model.ErrorResponse(err.Error(), nil))
    }

    // --- set cookie refresh token ---
    c.Cookie(&fiber.Cookie{
        Name:     "refreshToken",
        Value:    res.RefreshToken,
        HTTPOnly: true,
        Secure:   true,
        SameSite: "Strict",
        Path:     "/",
        Expires:  time.Now().Add(7 * 24 * time.Hour),
    })

    return c.JSON(model.SuccessResponse(res))
}


// REFRESH TOKEN
func (s *AuthService) RefreshToken(refreshToken string) (*model.LoginResponse, error) {

	userID, err := helper.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.authRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user inactive")
	}

	perms, _ := s.authRepo.GetUserPermissions(user.RoleID)

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


func (s *AuthService) HandleRefresh(c *fiber.Ctx) error {

    refreshToken := c.Cookies("refreshToken")
    if refreshToken == "" {
        return c.JSON(model.ErrorResponse("Missing refresh token", nil))
    }

    res, err := s.RefreshToken(refreshToken)
    if err != nil {
        c.ClearCookie("refreshToken")
        return c.JSON(model.ErrorResponse(err.Error(), nil))
    }

    // rotate new refresh token
    c.Cookie(&fiber.Cookie{
        Name:     "refreshToken",
        Value:    res.RefreshToken,
        HTTPOnly: true,
        Secure:   true,
        SameSite: "Strict",
        Path:     "/",
        Expires:  time.Now().Add(7 * 24 * time.Hour),
    })

    return c.JSON(model.SuccessResponse(res))
}

// LOGOUT
func (s *AuthService) Logout() error {
    // Karena refresh token tersimpan di cookie
    // logout hanya akan menghapus cookie di route
    return nil
}

func (s *AuthService) HandleLogout(c *fiber.Ctx) error {

    c.ClearCookie("refreshToken")
    s.Logout()

    return c.JSON(model.SuccessResponse(fiber.Map{
        "message": "Logged out",
    }))
}


// GetUserProfile mengambil data profil berdasarkan userID
func (s *AuthService) GetUserProfile(userID string) (*model.UserResponse, error) {

	user, err := s.authRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	perms, err := s.authRepo.GetUserPermissions(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load permissions")
	}

	return &model.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: perms,
	}, nil
}


func (s *AuthService) HandleGetProfile(c *fiber.Ctx) error {

    claims, ok := c.Locals("user").(*model.JWTClaims)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).
            JSON(model.ErrorResponse("Invalid user session", nil))
    }

    profile, err := s.GetUserProfile(claims.UserID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).
            JSON(model.ErrorResponse(err.Error(), nil))
    }

    return c.JSON(model.SuccessResponse(profile))
}
