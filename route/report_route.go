package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
	"go-fiber/middleware"
)

func SetupReportRoutes(app fiber.Router, svc *service.ReportService) {
	reports := app.Group("/reports",
		middleware.AuthMiddleware(),
	)

	reports.Get("/statistics",
		middleware.RequirePermission("achievement:read"),
		svc.Statistics,
	)

	reports.Get("/student/:id",
		middleware.RequirePermission("achievement:read"),
		svc.StudentStatistics,
	)
}
