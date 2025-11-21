package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/middleware"
)

// SetupAchievementRoutes contoh implementasi protected routes dengan RBAC
func SetupAchievementRoutes(app fiber.Router) {
	achievement := app.Group("/achievements")

	// GET /api/v1/achievements - List achievements
	// Memerlukan permission: achievement:read
	achievement.Get("/", 
		middleware.AuthMiddleware(), 
		middleware.RequirePermission("achievement:read"),
		func(c *fiber.Ctx) error {
			user := c.Locals("user").(*model.JWTClaims)
			
			return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]interface{}{
				"message": "List of achievements",
				"user":    user.Username,
				"role":    user.Role,
			}))
		},
	)

	// POST /api/v1/achievements - Create achievement
	// Memerlukan permission: achievement:create
	achievement.Post("/",
		middleware.AuthMiddleware(),
		middleware.RequirePermission("achievement:create"),
		func(c *fiber.Ctx) error {
			user := c.Locals("user").(*model.JWTClaims)

			return c.Status(fiber.StatusCreated).JSON(model.SuccessResponse(map[string]interface{}{
				"message":    "Achievement created successfully",
				"created_by": user.Username,
			}))
		},
	)

	// PUT /api/v1/achievements/:id - Update achievement
	// Memerlukan permission: achievement:update
	achievement.Put("/:id",
		middleware.AuthMiddleware(),
		middleware.RequirePermission("achievement:update"),
		func(c *fiber.Ctx) error {
			id := c.Params("id")
			user := c.Locals("user").(*model.JWTClaims)

			return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]interface{}{
				"message":    "Achievement updated successfully",
				"id":         id,
				"updated_by": user.Username,
			}))
		},
	)

	// DELETE /api/v1/achievements/:id - Delete achievement
	// Memerlukan permission: achievement:delete
	achievement.Delete("/:id",
		middleware.AuthMiddleware(),
		middleware.RequirePermission("achievement:delete"),
		func(c *fiber.Ctx) error {
			id := c.Params("id")
			user := c.Locals("user").(*model.JWTClaims)

			return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]interface{}{
				"message":    "Achievement deleted successfully",
				"id":         id,
				"deleted_by": user.Username,
			}))
		},
	)

	// POST /api/v1/achievements/:id/verify - Verify achievement
	// Memerlukan permission: achievement:verify (Dosen Wali)
	achievement.Post("/:id/verify",
		middleware.AuthMiddleware(),
		middleware.RequirePermission("achievement:verify"),
		func(c *fiber.Ctx) error {
			id := c.Params("id")
			user := c.Locals("user").(*model.JWTClaims)

			return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]interface{}{
				"message":     "Achievement verified successfully",
				"id":          id,
				"verified_by": user.Username,
				"role":        user.Role,
			}))
		},
	)

	// Contoh: Endpoint yang hanya bisa diakses Admin
	achievement.Get("/admin/stats",
		middleware.AuthMiddleware(),
		middleware.RequireRole("Admin"),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(map[string]interface{}{
				"message": "Admin statistics",
				"stats":   "All achievements statistics here",
			}))
		},
	)
}