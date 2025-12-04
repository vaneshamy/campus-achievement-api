package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
	"go-fiber/middleware"
)

func SetupLecturerRoutes(app fiber.Router, lecturerService *service.LecturerService) {

	lec := app.Group("/lecturers",
		middleware.AuthMiddleware(),
		middleware.RequireRole("Admin", "Dosen Wali"),
	)

	// GET /lecturers
	lec.Get("/", func(c *fiber.Ctx) error {
		return lecturerService.HandleGetLecturers(c)
	})

	// GET /lecturers/:id/advisees
	lec.Get("/:id/advisees", func(c *fiber.Ctx) error {
		return lecturerService.HandleGetAdvisees(c)
	})
}
