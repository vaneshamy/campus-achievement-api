package route

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "go-fiber/app/model"
    "go-fiber/app/service"
    "go-fiber/helper"
    "go-fiber/middleware"
)

func SetupAuthRoutes(app fiber.Router, authService *service.AuthService) {
    auth := app.Group("/auth")

    // LOGIN
    auth.Post("/login", func(c *fiber.Ctx) error {

        var req model.LoginRequest
        if err := c.BodyParser(&req); err != nil {
            return c.JSON(model.ErrorResponse("Invalid body", err))
        }

        // Validate input
        if errors := helper.ValidateLoginRequest(&req); len(errors) > 0 {
            return c.JSON(model.ErrorResponse("Validation failed", errors))
        }

        // Call service
        res, err := authService.Login(&req)
        if err != nil {
            return c.JSON(model.ErrorResponse(err.Error(), nil))
        }

        // Send refresh token via cookie
        c.Cookie(&fiber.Cookie{
            Name:     "refreshToken",
            Value:    res.RefreshToken,
            HTTPOnly: true,
            Secure:   true,
            SameSite: "Strict",
            Path:     "/",
            Expires:  time.Now().Add(7 * 24 * time.Hour),
        })

        return c.JSON(model.SuccessResponse(res))
    })

    // REFRESH TOKEN
    auth.Post("/refresh", func(c *fiber.Ctx) error {

        refreshToken := c.Cookies("refreshToken")
        if refreshToken == "" {
            return c.JSON(model.ErrorResponse("Missing refresh token", nil))
        }

        res, err := authService.RefreshToken(refreshToken)
        if err != nil {
            c.ClearCookie("refreshToken")
            return c.JSON(model.ErrorResponse(err.Error(), nil))
        }

        // Rotasi refresh token baru
        c.Cookie(&fiber.Cookie{
            Name:     "refreshToken",
            Value:    res.RefreshToken,
            HTTPOnly: true,
            Secure:   true,
            SameSite: "Strict",
            Path:     "/",
            Expires:  time.Now().Add(7 * 24 * time.Hour),
        })

        return c.JSON(model.SuccessResponse(res))
    })

    // LOGOUT
    auth.Post("/logout", func(c *fiber.Ctx) error {

        // Hapus cookie
        c.ClearCookie("refreshToken")

        // Ops: tidak ada logic di service
        authService.Logout()

        return c.JSON(model.SuccessResponse(map[string]string{
            "message": "Logged out",
        }))
    })
       
    auth.Get("/profile", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
        // Ambil claims user dari middleware
        claims, ok := c.Locals("user").(*model.JWTClaims)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse(
                "Invalid user session",
                nil,
            ))
        }

        // Ambil data user dari service
        profile, err := authService.GetUserProfile(claims.UserID)
        if err != nil {
            return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse(
                err.Error(),
                nil,
            ))
        }

        // Format response sesuai hanya untuk data yang dibutuhkan
        return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(profile))
    })

}
