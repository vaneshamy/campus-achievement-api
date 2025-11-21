package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go-fiber/helper"
)

// LoggerMiddleware mencatat setiap request
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request info
		duration := time.Since(start)
		helper.InfoLogger.Printf(
			"[%s] %s %s - Status: %d - Duration: %v",
			c.Method(),
			c.Path(),
			c.IP(),
			c.Response().StatusCode(),
			duration,
		)

		return err
	}
}