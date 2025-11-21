package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/app/service"
	"go-fiber/helper"
	"go-fiber/middleware"
)

// SetupAuthRoutes mendaftarkan routes untuk autentikasi
func SetupAuthRoutes(app fiber.Router, authService *service.AuthService) {
	auth := app.Group("/auth")

	// POST /api/v1/auth/login - Login endpoint
	auth.Post("/login", func(c *fiber.Ctx) error {
		// 1. Parse request body
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse(
				"Invalid request body",
				err.Error(),
			))
		}

		// 2. Validasi input
		if errors := helper.ValidateLoginRequest(&req); len(errors) > 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(model.ErrorResponse(
				"Validation failed",
				errors,
			))
		}

		// 3. Proses login
		response, err := authService.Login(&req)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				err.Error(),
				nil,
			))
		}

		// 4. Return response sukses
		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(response))
	})

	// POST /api/v1/auth/refresh - Refresh token endpoint
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		// 1. Parse request body
		var req model.RefreshTokenRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse(
				"Invalid request body",
				err.Error(),
			))
		}

		// 2. Validasi input
		if req.RefreshToken == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(model.ErrorResponse(
				"Refresh token is required",
				nil,
			))
		}

		// 3. Proses refresh token
		response, err := authService.RefreshToken(req.RefreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				err.Error(),
				nil,
			))
		}

		// 4. Return response sukses
		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(response))
	})

	// POST /api/v1/auth/logout - Logout endpoint (Protected)
	auth.Post("/logout", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		// 1. Parse request body
		var req model.RefreshTokenRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse(
				"Invalid request body",
				err.Error(),
			))
		}

		// 2. Proses logout (hapus refresh token)
		if err := authService.Logout(req.RefreshToken); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse(
				"Failed to logout",
				err.Error(),
			))
		}

		// 3. Return response sukses
		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]string{
			"message": "Logged out successfully",
		}))
	})

	// GET /api/v1/auth/profile - Get user profile endpoint (Protected)
	auth.Get("/profile", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		// 1. Ambil user dari context
		user := c.Locals("user").(*model.JWTClaims)

		// 2. Get user profile
		profile, err := authService.GetUserProfile(user.UserID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse(
				err.Error(),
				nil,
			))
		}

		// 3. Return response sukses
		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(profile))
	})
}
