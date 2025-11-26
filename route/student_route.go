package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
)

func SetupStudentRoutes(app fiber.Router, studentService *service.StudentService) {
	stu := app.Group("/students")

	stu.Get("/", func(c *fiber.Ctx) error {
		data, _ := studentService.GetAllStudents()
		return c.JSON(data)
	})

	stu.Get("/:id", func(c *fiber.Ctx) error {
		data, err := studentService.GetStudentByID(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON("student not found")
		}
		return c.JSON(data)
	})

	stu.Put("/:id/advisor", func(c *fiber.Ctx) error {
		var req struct{ AdvisorID string `json:"advisorId"` }
		c.BodyParser(&req)

		err := studentService.UpdateAdvisor(c.Params("id"), req.AdvisorID)
		if err != nil {
			return c.Status(400).JSON("failed")
		}
		return c.JSON("advisor updated")
	})
}
