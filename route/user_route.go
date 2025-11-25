package route

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/app/service"
	"go-fiber/middleware"
	"github.com/google/uuid"
)

func SetupUserRoutes(app fiber.Router, userService *service.UserService) {

	users := app.Group("/users", middleware.AuthMiddleware())

	users.Get("/", func(c *fiber.Ctx) error {
		result, err := userService.GetAllUsers()
		if err != nil {
			return c.JSON(model.ErrorResponse(err.Error(), nil))
		}

		return c.JSON(model.SuccessResponse(result))
	})

	users.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		result, err := userService.GetUserByID(id)
		if err != nil {
			return c.JSON(model.ErrorResponse(err.Error(), nil))
		}

		return c.JSON(model.SuccessResponse(result))
	})

	users.Post("/", func(c *fiber.Ctx) error {

		var req model.User
		if err := c.BodyParser(&req); err != nil {
			return c.JSON(model.ErrorResponse("Invalid body", err))
		}

		req.ID = uuid.New().String()

		if err := userService.CreateUser(&req); err != nil {
			return c.JSON(model.ErrorResponse(err.Error(), nil))
		}

		return c.JSON(model.SuccessResponse(fiber.Map{
			"message": "User created successfully",
			"id":      req.ID,
		}))
	})

	users.Put("/:id", func(c *fiber.Ctx) error {

		id := c.Params("id")
		var req model.User

		if err := c.BodyParser(&req); err != nil {
			return c.JSON(model.ErrorResponse("Invalid body", err))
		}

		if err := userService.UpdateUser(id, &req); err != nil {
			return c.JSON(model.ErrorResponse(err.Error(), nil))
		}

		return c.JSON(model.SuccessResponse(fiber.Map{
			"message": "User updated successfully",
		}))
	})

	users.Delete("/:id", func(c *fiber.Ctx) error {

		id := c.Params("id")

		if err := userService.DeleteUser(id); err != nil {
			return c.JSON(model.ErrorResponse(err.Error(), nil))
		}

		return c.JSON(model.SuccessResponse(fiber.Map{
			"message": "User deleted successfully",
		}))
	})

	users.Put("/:id/role", func(c *fiber.Ctx) error {

		id := c.Params("id")
		var body struct {
			RoleID string `json:"roleId"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.JSON(model.ErrorResponse("Invalid body", err))
		}

		if body.RoleID == "" {
			return c.JSON(model.ErrorResponse("roleId is required", nil))
		}

		if err := userService.AssignRole(id, body.RoleID); err != nil {
			return c.JSON(model.ErrorResponse(err.Error(), nil))
		}

		return c.JSON(model.SuccessResponse(fiber.Map{
			"message": "Role updated successfully",
		}))
	})
}
