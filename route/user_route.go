package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/service"
	"go-fiber/middleware"
)

func SetupUserRoutes(app fiber.Router, userService *service.UserService) {

	users := app.Group("/users",
		middleware.AuthMiddleware(),
		middleware.RequirePermission("user:manage"),
	)

	users.Get("/",
		userService.GetAllUsers,
	)

	users.Get("/:id",
		userService.GetUserByID,
	)

	users.Post("/",
		userService.CreateUser,
	)

	users.Put("/:id",
		userService.UpdateUser,
	)

	users.Delete("/:id",
		userService.DeleteUser,
	)

	users.Put("/:id/role",
		userService.AssignRole,
	)
}
