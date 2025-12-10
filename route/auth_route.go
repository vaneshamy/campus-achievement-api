package route

import (
    "github.com/gofiber/fiber/v2"
    "go-fiber/app/service"
    "go-fiber/middleware"
)

func SetupAuthRoutes(app fiber.Router, authService *service.AuthService) {

    auth := app.Group("/auth")

    // LOGIN
    auth.Post("/login", func(c *fiber.Ctx) error {
        return authService.HandleLogin(c)
    })

    // REFRESH
    auth.Post("/refresh", func(c *fiber.Ctx) error {
        return authService.HandleRefresh(c)
    })

    // LOGOUT
    auth.Post("/logout", func(c *fiber.Ctx) error {
        return authService.HandleLogout(c)
    })

    // PROFILE
    auth.Get("/profile", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
        return authService.HandleGetProfile(c)
    })
}
