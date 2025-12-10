package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
	"go-fiber/middleware"
)

func SetupLecturerRoutes(app fiber.Router, lecturerService *service.LecturerService) {

    lec := app.Group("/lecturers",
        middleware.AuthMiddleware(),
    )

    // GET /lecturers
    lec.Get("/",
        middleware.RequirePermission("user:manage"),
        lecturerService.GetLecturers,
    )

    // GET /lecturers/:id/advisees
    lec.Get("/:id/advisees",
        middleware.RequirePermission("user:manage"),
        lecturerService.GetAdvisees,
    )
}
