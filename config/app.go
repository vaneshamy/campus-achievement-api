package config

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go-fiber/app/repository"
	"go-fiber/app/service"
	"go-fiber/middleware"
	"go-fiber/route"
)

// NewApp membuat instance Fiber app dengan konfigurasi lengkap
func NewApp(db *sql.DB) *fiber.App {
	// Buat Fiber app
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Prestasi Backend API"),
		ErrorHandler: customErrorHandler,
	})

	// Setup middleware global
	app.Use(recover.New())                    // Recovery dari panic
	app.Use(middleware.CORSMiddleware())      // CORS
	app.Use(middleware.LoggerMiddleware())    // Logger

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo)

	// Setup routes
	api := app.Group("/api/v1")

	// Health check endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"message": "Server is running",
				"version": "1.0.0",
			},
		})
	})

	// Register route groups
	route.SetupAuthRoutes(api, authService)
	route.SetupAchievementRoutes(api) // Contoh protected routes

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Route not found",
		})
	})

	return app
}

// customErrorHandler menangani error secara custom
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": err.Error(),
	})
}
