package http

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/storex/go-crud/internal/transport/apierror"
)

func NewErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var fiberErr *fiber.Error
		var apiErr *apierror.Error
		code := fiber.StatusInternalServerError
		errorCode := "internal_error"
		message := "internal server error"

		if errors.As(err, &apiErr) {
			code = apiErr.StatusCode
			errorCode = toErrorCode(apiErr.Code)
			message = apiErr.Message
		} else if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			errorCode = toErrorCode(statusToErrorCode(code))
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
			"success": false,
			"error": fiber.Map{
				"code":        errorCode,
				"http_status": code,
				"message":     message,
				"data":        nil,
			},
			"meta": fiber.Map{
				"request_id": requestID,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			},
		})
	}
}

func statusToErrorCode(statusCode int) string {
	switch statusCode {
	case fiber.StatusBadRequest:
		return "bad_request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "forbidden"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusConflict:
		return "conflict"
	case fiber.StatusTooManyRequests:
		return "rate_limited"
	default:
		return "internal_error"
	}
}

func toErrorCode(raw string) string {
	up := strings.TrimSpace(raw)
	up = strings.ReplaceAll(up, "-", "_")
	return strings.ToUpper(up)
}
