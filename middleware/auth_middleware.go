package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/helper"
)

// AuthMiddleware memvalidasi JWT token
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ekstrak token dari header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				"Missing authorization header",
				nil,
			))
		}

		// 2. Validasi format Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				"Invalid authorization format",
				nil,
			))
		}

		tokenString := parts[1]

		// 3. Validasi token
		claims, err := helper.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				"Invalid or expired token",
				err.Error(),
			))
		}

		// 4. Simpan claims ke context untuk digunakan di handler
		c.Locals("user", claims)

		return c.Next()
	}
}

// RequirePermission middleware untuk cek permission spesifik
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("user").(*model.JWTClaims)

		for _, p := range claims.Permissions {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse(
			"Akses ditolak: permission tidak mencukupi",
			map[string]string{
				"required": permission,
				"role": claims.Role,
			},
		))
	}
}


// RequireRole middleware untuk cek role spesifik
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil user claims dari context
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				"Unauthorized",
				nil,
			))
		}

		claims, ok := user.(*model.JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
				"Invalid user context",
				nil,
			))
		}

		// 2. Cek apakah user memiliki role yang diperlukan
		hasRole := false
		for _, role := range roles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse(
				"Insufficient role",
				map[string]interface{}{
					"required":   roles,
					"user_role":  claims.Role,
				},
			))
		}

		return c.Next()
	}
}