package middleware

import (
	"time"
    "fmt"

	"github.com/gofiber/fiber/v2"
	"go-fiber/helper"
)

// LoggerMiddleware mencatat setiap request
func LoggerMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        err := c.Next()  // tetap ambil err, supaya tidak unused

        duration := time.Since(start)

        fmt.Println("DEBUG RAW:", c.Response().StatusCode(), "FINAL:", c.Context().Response.StatusCode())

        status := c.Context().Response.StatusCode()

        helper.InfoLogger.Printf(
            "[%s] %s %s - Status: %d - Duration: %v",
            c.Method(),
            c.Path(),
            c.IP(),
            status,
            duration,
        )

        return err // tetap kembalikan err, supaya tidak unused
    }
}
