package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/storex/go-crud/internal/app"
	"github.com/storex/go-crud/internal/config"
	"github.com/storex/go-crud/internal/platform/async"
	"github.com/storex/go-crud/internal/platform/db"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer pool.Close()

	auditLogger := async.NewAuditLogger(cfg.AuditBufferSize, cfg.AuditWorkerGoroutines, logger)
	auditLogger.Start(ctx)
	defer auditLogger.Stop()

	server := app.NewServer(cfg, logger, pool, auditLogger)
	if err := app.Run(ctx, cfg, logger, server); err != nil {
		logger.Error("server stopped", "err", err)
	}
}
