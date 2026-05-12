package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/storex/go-crud/internal/service"
	"github.com/storex/go-crud/internal/transport/http/handler"
	"github.com/storex/go-crud/internal/transport/http/middleware"
)

// APIV1Prefix is the URL prefix for all versioned JSON API routes.
const APIV1Prefix = "/api/v1"

func RegisterRoutes(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	itemHandler *handler.ItemHandler,
	userHandler *handler.UserHandler,
	brandHandler *handler.BrandHandler,
	healthHandler *handler.HealthHandler,
	authService *service.AuthService,
) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":      "go-crud-api",
			"swagger_ui":   SwaggerUIPath,
			"openapi_spec": OpenAPISpecPath,
			"api_base":     APIV1Prefix,
			"version":      "v1",
		})
	})
	app.Get("/healthz", healthHandler.Liveness)
	app.Get("/readyz", healthHandler.Readiness)

	v1 := app.Group(APIV1Prefix)
	v1.Post("/auth/admin/login", authHandler.AdminLogin)
	v1.Post("/auth/admin/refresh", authHandler.AdminRefresh)
	v1.Post("/users/login", userHandler.Login)
	v1.Post("/users/register", userHandler.Register)
	v1.Post("/users/refresh", userHandler.Refresh)

	protected := v1.Group("", middleware.Authenticate(authService))
	protected.Get("/items", middleware.Authorize(service.RoleAdmin, service.RoleUser), itemHandler.List)
	protected.Get("/items/:id", middleware.Authorize(service.RoleAdmin, service.RoleUser), itemHandler.GetByID)
	protected.Post("/items", middleware.Authorize(service.RoleAdmin), itemHandler.Create)
	protected.Get("/users/me", middleware.Authorize(service.RoleUser), userHandler.GetMyProfile)
	protected.Patch("/users/me", middleware.Authorize(service.RoleUser), userHandler.UpdateMyProfile)
	protected.Get("/brands", middleware.Authorize(service.RoleAdmin, service.RoleUser), brandHandler.List)
	protected.Post("/brands", middleware.Authorize(service.RoleAdmin), brandHandler.Create)
}
