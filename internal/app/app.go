package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/storex/go-crud/internal/config"
	"github.com/storex/go-crud/internal/platform/async"
	repopg "github.com/storex/go-crud/internal/repository/postgres"
	"github.com/storex/go-crud/internal/service"
	httptransport "github.com/storex/go-crud/internal/transport/http"
	"github.com/storex/go-crud/internal/transport/http/handler"
	observemw "github.com/storex/go-crud/internal/transport/http/middleware"
)

func NewServer(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool, auditLogger *async.AuditLogger) *fiber.App {
	itemRepo := repopg.NewItemRepository(pool)
	itemService := service.NewItemService(itemRepo, auditLogger)
	authService := service.NewAuthService(
		cfg.JWTSecret,
		cfg.AdminEmail,
		cfg.AdminPassword,
		cfg.UserEmail,
		cfg.UserPassword,
	)

	authHandler := handler.NewAuthHandler(authService)
	itemHandler := handler.NewItemHandler(itemService)
	healthHandler := handler.NewHealthHandler(pool)

	app := fiber.New(fiber.Config{
		ErrorHandler:      httptransport.NewErrorHandler(logger),
		DisableKeepalive:  false,
		EnablePrintRoutes: false,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		AppName:           "go-crud-api",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://127.0.0.1:8080",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))
	app.Use(observemw.RequestIdentity())
	app.Use(observemw.RequestLogger(logger))

	httptransport.RegisterRoutes(app, authHandler, itemHandler, healthHandler, authService)
	httptransport.RegisterSwaggerRoutes(app)

	return app
}

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger, app *fiber.App) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", "err", err)
		}
	}()

	logger.Info(
		"api started",
		"addr", cfg.HTTPAddr,
		"env", cfg.AppEnv,
		"swagger_ui", "http://127.0.0.1"+cfg.HTTPAddr+"/swagger",
		"openapi", "http://127.0.0.1"+cfg.HTTPAddr+"/openapi.yaml",
	)
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}
