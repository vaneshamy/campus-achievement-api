package route

import (
    "go-fiber/app/service"
    "go-fiber/middleware"

    "github.com/gofiber/fiber/v2"
)

func SetupStudentRoutes(
    router fiber.Router, 
    studentService *service.StudentService,
    achievementService *service.AchievementService,
) {


    student := router.Group("/students",
        middleware.AuthMiddleware(),
        middleware.RequirePermission("user:manage"), 
    )

    student.Get("/",
        studentService.GetAllStudentsService,
    )

    student.Get("/:id",
        studentService.GetStudentDetailService,
    )

    student.Put("/:id/advisor",
        studentService.UpdateAdvisorService,
    )

     student.Get("/:id/achievements",
        achievementService.GetStudentAchievements,
    )
}
