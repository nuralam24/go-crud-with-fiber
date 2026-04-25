package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

func RequestIdentity() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Locals("request_id", requestID)
		c.Set(requestIDHeader, requestID)
		return c.Next()
	}
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		requestID, _ := c.Locals("request_id").(string)
		logger.Info(
			"http-request",
			"request_id", requestID,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", latency.Milliseconds(),
			"ip", c.IP(),
		)

		return err
	}
}
