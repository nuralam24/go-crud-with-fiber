package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/http/handler"
	"github.com/storex/go-crud/internal/transport/http/middleware"
)

func RegisterRoutes(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	itemHandler *handler.ItemHandler,
	healthHandler *handler.HealthHandler,
	authService *service.AuthService,
) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":      "go-crud-api",
			"swagger_ui":   "/swagger",
			"openapi_spec": "/openapi.yaml",
			"version":      "v1",
		})
	})
	app.Get("/healthz", healthHandler.Liveness)
	app.Get("/readyz", healthHandler.Readiness)

	v1 := app.Group("/v1")
	v1.Post("/auth/login", authHandler.Login)

	protected := v1.Group("", middleware.Authenticate(authService))
	protected.Get("/items", middleware.Authorize(service.RoleAdmin, service.RoleUser), itemHandler.List)
	protected.Get("/items/:id", middleware.Authorize(service.RoleAdmin, service.RoleUser), itemHandler.GetByID)
	protected.Post("/items", middleware.Authorize(service.RoleAdmin), itemHandler.Create)
}
