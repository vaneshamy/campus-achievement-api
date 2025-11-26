package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
)

func SetupLecturerRoutes(app fiber.Router, lecturerService *service.LecturerService) {
	lec := app.Group("/lecturers")

	lec.Get("/", func(c *fiber.Ctx) error {
		data, _ := lecturerService.GetLecturers()
		return c.JSON(data)
	})

	lec.Get("/:id/advisees", func(c *fiber.Ctx) error {
		data, _ := lecturerService.GetAdvisees(c.Params("id"))
		return c.JSON(data)
	})
}
