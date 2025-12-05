package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
	"go-fiber/middleware"
)

func SetupAchievementRoutes(app fiber.Router, svc *service.AchievementService) {

	ach := app.Group("/achievements",
		middleware.AuthMiddleware(),
		middleware.RequireRole("Mahasiswa", "Dosen Wali", "Admin"),
	)

	ach.Get("/", svc.List)
	ach.Get("/:id", svc.Detail)
	ach.Post("/", middleware.RequireRole("Mahasiswa"), svc.Create)
	ach.Put("/:id", middleware.RequireRole("Mahasiswa"), svc.Update)
	ach.Delete("/:id", middleware.RequireRole("Mahasiswa"), svc.Delete)
	ach.Post("/:id/submit", middleware.RequireRole("Mahasiswa"), svc.Submit)
	ach.Post("/:id/verify",
		middleware.RequireRole("Dosen Wali", "Admin"),
		svc.Verify,
	)
	ach.Post("/:id/reject",
		middleware.RequireRole("Dosen Wali", "Admin"),
		svc.Reject,
	)
	ach.Get("/:id/history", svc.History)
	ach.Post("/:id/attachments", svc.UploadAttachment)
}
