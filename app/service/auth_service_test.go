package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go-fiber/app/model"
	"go-fiber/app/service"
	"go-fiber/helper"
)

// MockAuthRepository adalah mock untuk AuthRepository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) FindByUsernameOrEmail(identifier string) (*model.User, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) FindByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserPermissions(roleID string) ([]string, error) {
	args := m.Called(roleID)
	return args.Get(0).([]string), args.Error(1)
}

// Helper function untuk membuat test user
func createTestUser() *model.User {
	hashedPassword, _ := helper.HashPassword("password123")
	return &model.User{
		ID:           "user-123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		FullName:     "Test User",
		RoleID:       "role-1",
		RoleName:     "Admin",
		IsActive:     true,
	}
}

// Test Login Success
func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	permissions := []string{"read:users", "write:users"}

	mockRepo.On("FindByUsernameOrEmail", "testuser").Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return(permissions, nil)

	req := &model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	res, err := authService.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, testUser.ID, res.User.ID)
	assert.Equal(t, testUser.Username, res.User.Username)
	assert.Equal(t, testUser.FullName, res.User.FullName)
	assert.Equal(t, testUser.RoleName, res.User.Role)
	assert.Equal(t, permissions, res.User.Permissions)

	mockRepo.AssertExpectations(t)
}

// Test Login - User Not Found
func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	mockRepo.On("FindByUsernameOrEmail", "wronguser").Return(nil, errors.New("user not found"))

	req := &model.LoginRequest{
		Username: "wronguser",
		Password: "password123",
	}

	res, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid username or password")

	mockRepo.AssertExpectations(t)
}

// Test Login - Wrong Password
func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	mockRepo.On("FindByUsernameOrEmail", "testuser").Return(testUser, nil)

	req := &model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	res, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid username or password")

	mockRepo.AssertExpectations(t)
}

// Test Login - Inactive User
func TestLogin_InactiveUser(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	testUser.IsActive = false

	mockRepo.On("FindByUsernameOrEmail", "testuser").Return(testUser, nil)

	req := &model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	res, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "user not active")

	mockRepo.AssertExpectations(t)
}

// Test HandleLogin - Success
func TestHandleLogin_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	permissions := []string{"read:users"}

	mockRepo.On("FindByUsernameOrEmail", "testuser").Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return(permissions, nil)

	app := fiber.New()
	app.Post("/login", authService.HandleLogin)

	reqBody := model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response body
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result["success"].(bool))

	// Verify cookie is set
	cookies := resp.Cookies()
	var refreshTokenFound bool
	for _, cookie := range cookies {
		if cookie.Name == "refreshToken" {
			refreshTokenFound = true
			assert.NotEmpty(t, cookie.Value)
			assert.True(t, cookie.HttpOnly)
			assert.True(t, cookie.Secure)
			break
		}
	}
	assert.True(t, refreshTokenFound, "refreshToken cookie should be set")

	mockRepo.AssertExpectations(t)
}

// Test HandleLogin - Invalid Body
func TestHandleLogin_InvalidBody(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	app := fiber.New()
	app.Post("/login", authService.HandleLogin)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result["success"].(bool))
}

// Test RefreshToken - Success
func TestRefreshToken_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	permissions := []string{"read:users"}

	// Generate valid refresh token
	refreshToken, _ := helper.GenerateRefreshToken(testUser.ID)

	mockRepo.On("FindByID", testUser.ID).Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return(permissions, nil)

	res, err := authService.RefreshToken(refreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, testUser.ID, res.User.ID)

	mockRepo.AssertExpectations(t)
}

// Test RefreshToken - Invalid Token
func TestRefreshToken_InvalidToken(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	res, err := authService.RefreshToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, res)
}

// Test RefreshToken - User Not Found
func TestRefreshToken_UserNotFound(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	refreshToken, _ := helper.GenerateRefreshToken("nonexistent-user")

	mockRepo.On("FindByID", "nonexistent-user").Return(nil, errors.New("user not found"))

	res, err := authService.RefreshToken(refreshToken)

	assert.Error(t, err)
	assert.Nil(t, res)

	mockRepo.AssertExpectations(t)
}

// Test RefreshToken - Inactive User
func TestRefreshToken_InactiveUser(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	testUser.IsActive = false

	refreshToken, _ := helper.GenerateRefreshToken(testUser.ID)

	mockRepo.On("FindByID", testUser.ID).Return(testUser, nil)

	res, err := authService.RefreshToken(refreshToken)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "user inactive")

	mockRepo.AssertExpectations(t)
}

// Test HandleRefresh - Success
func TestHandleRefresh_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	permissions := []string{"read:users"}
	refreshToken, _ := helper.GenerateRefreshToken(testUser.ID)

	mockRepo.On("FindByID", testUser.ID).Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return(permissions, nil)

	app := fiber.New()
	app.Post("/refresh", authService.HandleRefresh)

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Cookie", "refreshToken="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result["success"].(bool))

	mockRepo.AssertExpectations(t)
}

// Test HandleRefresh - Missing Token
func TestHandleRefresh_MissingToken(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	app := fiber.New()
	app.Post("/refresh", authService.HandleRefresh)

	req := httptest.NewRequest("POST", "/refresh", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "Missing refresh token")
}

// Test HandleLogout
func TestHandleLogout_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	app := fiber.New()
	app.Post("/logout", authService.HandleLogout)

	req := httptest.NewRequest("POST", "/logout", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result["success"].(bool))

	// Check that refreshToken cookie is cleared
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refreshToken" {
			// Cookie should be expired or have empty value
			assert.True(t, cookie.MaxAge < 0 || cookie.Value == "")
		}
	}
}

// Test GetUserProfile - Success
func TestGetUserProfile_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	permissions := []string{"read:users", "write:users"}

	mockRepo.On("FindByID", testUser.ID).Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return(permissions, nil)

	profile, err := authService.GetUserProfile(testUser.ID)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, testUser.ID, profile.ID)
	assert.Equal(t, testUser.Username, profile.Username)
	assert.Equal(t, testUser.FullName, profile.FullName)
	assert.Equal(t, testUser.RoleName, profile.Role)
	assert.Equal(t, permissions, profile.Permissions)

	mockRepo.AssertExpectations(t)
}

// Test GetUserProfile - User Not Found
func TestGetUserProfile_UserNotFound(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	mockRepo.On("FindByID", "nonexistent").Return(nil, errors.New("user not found"))

	profile, err := authService.GetUserProfile("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "user not found")

	mockRepo.AssertExpectations(t)
}

// Test GetUserProfile - Permission Load Failed
func TestGetUserProfile_PermissionLoadFailed(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()

	mockRepo.On("FindByID", testUser.ID).Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return([]string{}, errors.New("db error"))

	profile, err := authService.GetUserProfile(testUser.ID)

	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "failed to load permissions")

	mockRepo.AssertExpectations(t)
}

// Test HandleGetProfile - Success
func TestHandleGetProfile_Success(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	testUser := createTestUser()
	permissions := []string{"read:users"}

	mockRepo.On("FindByID", testUser.ID).Return(testUser, nil)
	mockRepo.On("GetUserPermissions", testUser.RoleID).Return(permissions, nil)

	app := fiber.New()
	app.Get("/profile", func(c *fiber.Ctx) error {
		// Simulate authenticated user
		c.Locals("user", &model.JWTClaims{
			UserID:      testUser.ID,
			Username:    testUser.Username,
			Role:        testUser.RoleName,
			Permissions: permissions,
		})
		return authService.HandleGetProfile(c)
	})

	req := httptest.NewRequest("GET", "/profile", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result["success"].(bool))

	mockRepo.AssertExpectations(t)
}

// Test HandleGetProfile - Invalid Session
func TestHandleGetProfile_InvalidSession(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	authService := service.NewAuthService(mockRepo)

	app := fiber.New()
	app.Get("/profile", authService.HandleGetProfile)

	req := httptest.NewRequest("GET", "/profile", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "Invalid user session")
}