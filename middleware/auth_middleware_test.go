package middleware_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"go-fiber/app/model"
	"go-fiber/helper"
	"go-fiber/middleware"
)

// Test AuthMiddleware - Success
func TestAuthMiddleware_Success(t *testing.T) {
	app := fiber.New()

	// Create test user
	testUser := &model.User{
		ID:       "user-123",
		Username: "testuser",
		FullName: "Test User",
		RoleID:   "role-1",
		RoleName: "Admin",
	}

	permissions := []string{"read:users", "write:users"}

	// Generate valid access token
	token, err := helper.GenerateAccessToken(testUser, permissions)
	assert.NoError(t, err)

	app.Get("/protected", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*model.JWTClaims)
		return c.JSON(fiber.Map{
			"user_id": claims.UserID,
			"username": claims.Username,
		})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, testUser.ID, result["user_id"])
	assert.Equal(t, testUser.Username, result["username"])
}

// Test AuthMiddleware - Missing Authorization Header
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Missing authorization header")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test AuthMiddleware - Invalid Format (No Bearer)
func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidToken123")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Invalid authorization format")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test AuthMiddleware - Invalid Token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Invalid or expired token")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test RequirePermission - Has Permission
func TestRequirePermission_HasPermission(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", func(c *fiber.Ctx) error {
		// Simulate authenticated user with permission
		c.Locals("user", &model.JWTClaims{
			UserID:      "user-123",
			Username:    "admin",
			Role:        "Admin",
			Permissions: []string{"read:users", "write:users", "delete:users"},
		})
		return c.Next()
	}, middleware.RequirePermission("write:users"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Access granted", result["message"])
}

// Test RequirePermission - Missing Permission
func TestRequirePermission_MissingPermission(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", func(c *fiber.Ctx) error {
		// Simulate authenticated user without required permission
		c.Locals("user", &model.JWTClaims{
			UserID:      "user-123",
			Username:    "user",
			Role:        "User",
			Permissions: []string{"read:users"},
		})
		return c.Next()
	}, middleware.RequirePermission("delete:users"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Akses ditolak")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test RequireRole - Has Role
func TestRequireRole_HasRole(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", func(c *fiber.Ctx) error {
		// Simulate authenticated admin user
		c.Locals("user", &model.JWTClaims{
			UserID:      "user-123",
			Username:    "admin",
			Role:        "Admin",
			Permissions: []string{"read:users"},
		})
		return c.Next()
	}, middleware.RequireRole("Admin", "SuperAdmin"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Access granted", result["message"])
}

// Test RequireRole - Missing Role
func TestRequireRole_MissingRole(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", func(c *fiber.Ctx) error {
		// Simulate authenticated user with wrong role
		c.Locals("user", &model.JWTClaims{
			UserID:      "user-123",
			Username:    "user",
			Role:        "User",
			Permissions: []string{"read:users"},
		})
		return c.Next()
	}, middleware.RequireRole("Admin", "SuperAdmin"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Insufficient role")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test RequireRole - Nil User Context
func TestRequireRole_NilUserContext(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", middleware.RequireRole("Admin"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Unauthorized")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test RequireRole - Invalid User Context Type
func TestRequireRole_InvalidUserContextType(t *testing.T) {
	app := fiber.New()

	app.Get("/admin", func(c *fiber.Ctx) error {
		// Set invalid user context type
		c.Locals("user", "invalid-type")
		return c.Next()
	}, middleware.RequireRole("Admin"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Access granted"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	
	if success, ok := result["success"].(bool); ok {
		assert.False(t, success)
		assert.Contains(t, result["message"], "Invalid user context")
	} else {
		t.Fatalf("Expected 'success' field in response, got: %+v", result)
	}
}

// Test Middleware Chain - Auth + Permission
func TestMiddlewareChain_AuthAndPermission(t *testing.T) {
	app := fiber.New()

	testUser := &model.User{
		ID:       "user-123",
		Username: "testuser",
		RoleID:   "role-1",
		RoleName: "Admin",
	}

	permissions := []string{"read:users", "write:users"}
	token, _ := helper.GenerateAccessToken(testUser, permissions)

	app.Get("/admin-action",
		middleware.AuthMiddleware(),
		middleware.RequirePermission("write:users"),
		func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "Action performed"})
		},
	)

	req := httptest.NewRequest("GET", "/admin-action", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Action performed", result["message"])
}

// Test Middleware Chain - Auth + Role
func TestMiddlewareChain_AuthAndRole(t *testing.T) {
	app := fiber.New()

	testUser := &model.User{
		ID:       "user-123",
		Username: "admin",
		RoleID:   "role-1",
		RoleName: "Admin",
	}

	permissions := []string{"read:users"}
	token, _ := helper.GenerateAccessToken(testUser, permissions)

	app.Get("/admin-panel",
		middleware.AuthMiddleware(),
		middleware.RequireRole("Admin"),
		func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "Welcome to admin panel"})
		},
	)

	req := httptest.NewRequest("GET", "/admin-panel", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Welcome to admin panel", result["message"])
}

// Test Middleware Chain - Auth + Role + Permission
func TestMiddlewareChain_Full(t *testing.T) {
	app := fiber.New()

	testUser := &model.User{
		ID:       "user-123",
		Username: "admin",
		RoleID:   "role-1",
		RoleName: "Admin",
	}

	permissions := []string{"read:users", "write:users", "delete:users"}
	token, _ := helper.GenerateAccessToken(testUser, permissions)

	app.Delete("/user/:id",
		middleware.AuthMiddleware(),
		middleware.RequireRole("Admin", "SuperAdmin"),
		middleware.RequirePermission("delete:users"),
		func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"message": "User deleted",
				"id":      c.Params("id"),
			})
		},
	)

	req := httptest.NewRequest("DELETE", "/user/456", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "User deleted", result["message"])
	assert.Equal(t, "456", result["id"])
}