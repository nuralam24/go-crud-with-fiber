package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/storex/go-crud/internal/config"
)

func NewPostgresPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PGConnString)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	// Transaction-pool friendly for PgBouncer.
	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	poolCfg.MaxConns = cfg.PGMaxConns
	poolCfg.MinConns = cfg.PGMinConns
	poolCfg.MaxConnLifetime = cfg.PGMaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.PGMaxConnIdleTime
	poolCfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("new pg pool: %w", err)
	}

	var pingErr error
	for attempt := 1; attempt <= cfg.PGPingRetries; attempt++ {
		// Remote managed Postgres can have temporary cold-start/network delay.
		pingCtx, cancel := context.WithTimeout(ctx, cfg.PGPingTimeout)
		pingErr = pool.Ping(pingCtx)
		cancel()

		if pingErr == nil {
			return pool, nil
		}

		if attempt < cfg.PGPingRetries {
			select {
			case <-ctx.Done():
				pool.Close()
				return nil, fmt.Errorf("ping pg canceled: %w", ctx.Err())
			case <-time.After(cfg.PGPingRetryDelay):
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf("ping pg failed after %d attempts: %w", cfg.PGPingRetries, pingErr)
}
