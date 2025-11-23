package route

import (
	"time"

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

		// 3. Proses login via service
		response, err := authService.Login(&req)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				err.Error(),
				nil,
			))
		}

		// 4. Simpan refresh token di HTTPOnly cookie
		refreshCookie := &fiber.Cookie{
			Name:     "refreshToken",
			Value:    response.RefreshToken,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/",
		}

		// Set expiry sesuai ENV (default 7 hari)
		if dur, err := time.ParseDuration(helper.GetEnvOrDefault("JWT_REFRESH_EXPIRES_IN", "168h")); err == nil {
			refreshCookie.Expires = time.Now().Add(dur)
		} else {
			refreshCookie.Expires = time.Now().Add(7 * 24 * time.Hour)
		}

		c.Cookie(refreshCookie)

		// 5. Return response (refresh token ikut dikirim)
		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(response))
	})

	// POST /api/v1/auth/refresh - Refresh access token
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		// Ambil refresh token dari cookie
		refreshToken := c.Cookies("refreshToken")
		if refreshToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				"Refresh token missing",
				nil,
			))
		}

		response, err := authService.RefreshToken(refreshToken)
		if err != nil {
			// Invalid -> hapus cookie
			c.Cookie(&fiber.Cookie{
				Name:     "refreshToken",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Strict",
				Path:     "/",
			})

			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				err.Error(),
				nil,
			))
		}

		// Rotasi refresh token baru (jika ada)
		if response.RefreshToken != "" {
			newCookie := &fiber.Cookie{
				Name:     "refreshToken",
				Value:    response.RefreshToken,
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Strict",
				Path:     "/",
			}

			if dur, err := time.ParseDuration(helper.GetEnvOrDefault("JWT_REFRESH_EXPIRES_IN", "168h")); err == nil {
				newCookie.Expires = time.Now().Add(dur)
			} else {
				newCookie.Expires = time.Now().Add(7 * 24 * time.Hour)
			}

			c.Cookie(newCookie)
		}

		// Return response lengkap (token + refresh token)
		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(response))
	})

	// POST /api/v1/auth/logout - Logout endpoint
	auth.Post("/logout", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		// Clear cookie
		c.Cookie(&fiber.Cookie{
			Name:     "refreshToken",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
			Path:     "/",
		})

		// Tetap panggil service agar kontrak terjaga
		_ = authService.Logout("")

		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]string{
			"message": "Logged out successfully",
		}))
	})

	// GET /api/v1/auth/profile - Profil user (Protected)
	auth.Get("/profile", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		userClaims := c.Locals("user").(*model.JWTClaims)

		profile, err := authService.GetUserProfile(userClaims.UserID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse(
				err.Error(),
				nil,
			))
		}

		return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(profile))
	})
}
