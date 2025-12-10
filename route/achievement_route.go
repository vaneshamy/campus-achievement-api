package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
	"go-fiber/middleware"
)

func SetupAchievementRoutes(app fiber.Router, svc *service.AchievementService) {

	ach := app.Group("/achievements",
		middleware.AuthMiddleware(),
	)
	ach.Get("/",
		middleware.RequirePermission("achievement:view"),
		svc.List,
	)
	ach.Get("/:id",
		middleware.RequirePermission("achievement:read"),
		svc.Detail,
	)
	ach.Post("/",
		middleware.RequirePermission("achievement:create"),
		svc.Create,
	)
	ach.Put("/:id",
		middleware.RequirePermission("achievement:update"),
		svc.Update,
	)
	ach.Delete("/:id",
		middleware.RequirePermission("achievement:delete"),
		svc.Delete,
	)
	ach.Post("/:id/submit",
		middleware.RequirePermission("achievement:update"),
		svc.Submit,
	)
	ach.Post("/:id/verify",
		middleware.RequirePermission("achievement:verify"),
		svc.Verify,
	)
	ach.Post("/:id/reject",
		middleware.RequirePermission("achievement:reject"),
		svc.Reject,
	)
	ach.Get("/:id/history",
		middleware.RequirePermission("achievement:read"),
		svc.History,
	)
	ach.Post("/:id/attachments",
		middleware.RequirePermission("achievement:update"),
		svc.UploadAttachment,
	)
}
