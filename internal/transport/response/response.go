package response

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type Meta struct {
	RequestID  string      `json:"request_id"`
	Timestamp  string      `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func Single(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    buildMeta(c, nil),
	})
}

func List(c *fiber.Ctx, status int, data interface{}, pagination Pagination) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    buildMeta(c, &pagination),
	})
}

func buildMeta(c *fiber.Ctx, pagination *Pagination) Meta {
	requestID, _ := c.Locals("request_id").(string)
	return Meta{
		RequestID:  requestID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Pagination: pagination,
	}
}
