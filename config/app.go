package config

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"go-fiber/app/repository"
	"go-fiber/app/service"
	"go-fiber/middleware"
	"go-fiber/route"
)

/*
	NewApp membuat instance Fiber app dengan konfigurasi lengkap
*/
func NewApp(db *sql.DB) *fiber.App {

	// Buat Fiber app
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Prestasi Backend API"),
		ErrorHandler: customErrorHandler,
	})

	// Middleware global
	app.Use(recover.New())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.LoggerMiddleware())

	// PostgreSQL repositories
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	achievementRepo := repository.NewAchievementRepository(db)

	// Mongo
	mongoClient, err := NewMongoClient()
	if err != nil {
		log.Fatal("❌ Failed to connect MongoDB:", err)
	}

	mongoDB := GetMongoDatabase(mongoClient)
	achievementsColl := mongoDB.Collection("achievements")

	mongoAchievementRepo := repository.NewMongoAchievementRepository(achievementsColl)

	// service
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo, studentRepo, lecturerRepo)
	studentService := service.NewStudentService(studentRepo)
	lecturerService := service.NewLecturerService(lecturerRepo)

	achievementService := service.NewAchievementService(
		achievementRepo,
		mongoAchievementRepo,
		studentRepo,
	)

	// route
	api := app.Group("/api/v1")

	// Health Check
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
	route.SetupAchievementRoutes(api, achievementService)
	route.SetupUserRoutes(api, userService)
	route.SetupStudentRoutes(api, studentService, achievementService)
	route.SetupLecturerRoutes(api, lecturerService)

	// 404 Handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Route not found",
		})
	})

	return app
}

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
