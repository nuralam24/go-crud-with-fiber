package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "database is unavailable")
	}

	return c.JSON(fiber.Map{
		"status": "ready",
	})
}
