package http

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func NewErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var fiberErr *fiber.Error
		code := fiber.StatusInternalServerError
		message := "internal server error"

		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			message = fiberErr.Message
		}

		requestID, _ := c.Locals("request_id").(string)
		if code >= fiber.StatusInternalServerError {
			logger.Error(
				"http-request-failed",
				"request_id", requestID,
				"method", c.Method(),
				"path", c.Path(),
				"status", code,
				"error", err.Error(),
			)
		}

		return c.Status(code).JSON(fiber.Map{
			"error":      message,
			"request_id": requestID,
		})
	}
}
