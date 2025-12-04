package route

import (
    "go-fiber/app/service"
    "go-fiber/middleware"

    "github.com/gofiber/fiber/v2"
)

func SetupStudentRoutes(router fiber.Router, studentService *service.StudentService) {

    student := router.Group("/students",
        middleware.AuthMiddleware(),
        middleware.RequireRole("Admin"),
    )

    student.Get("/", studentService.GetAllStudentsService)
    student.Get("/:id", studentService.GetStudentDetailService)
    student.Put("/:id/advisor", studentService.UpdateAdvisorService)
}
