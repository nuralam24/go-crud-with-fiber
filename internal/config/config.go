package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv                string
	HTTPAddr              string
	ShutdownTimeout       time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	PGConnString          string
	PGPingTimeout         time.Duration
	PGPingRetries         int
	PGPingRetryDelay      time.Duration
	PGMaxConns            int32
	PGMinConns            int32
	PGMaxConnLifetime     time.Duration
	PGMaxConnIdleTime     time.Duration
	JWTSecret             string
	AdminEmail            string
	AdminPassword         string
	UserEmail             string
	UserPassword          string
	AuditBufferSize       int
	AuditWorkerGoroutines int
}

func Load() Config {
	cfg := Config{
		AppEnv:                getString("APP_ENV", "dev"),
		HTTPAddr:              getString("HTTP_ADDR", ":8080"),
		ShutdownTimeout:       getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		ReadTimeout:           getDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:          getDuration("WRITE_TIMEOUT", 5*time.Second),
		IdleTimeout:           getDuration("IDLE_TIMEOUT", 60*time.Second),
		PGConnString:          mustGet("PG_DSN"),
		PGPingTimeout:         getDuration("PG_PING_TIMEOUT", 15*time.Second),
		PGPingRetries:         getInt("PG_PING_RETRIES", 5),
		PGPingRetryDelay:      getDuration("PG_PING_RETRY_DELAY", 2*time.Second),
		PGMaxConns:            int32(getInt("PG_MAX_CONNS", 120)),
		PGMinConns:            int32(getInt("PG_MIN_CONNS", 20)),
		PGMaxConnLifetime:     getDuration("PG_MAX_CONN_LIFETIME", 15*time.Minute),
		PGMaxConnIdleTime:     getDuration("PG_MAX_CONN_IDLE_TIME", 3*time.Minute),
		JWTSecret:             mustGet("JWT_SECRET"),
		AdminEmail:            getString("ADMIN_EMAIL", "admin@gmail.com"),
		AdminPassword:         getString("ADMIN_PASSWORD", "12345"),
		UserEmail:             getString("USER_EMAIL", "user@gmail.com"),
		UserPassword:          getString("USER_PASSWORD", "12345"),
		AuditBufferSize:       getInt("AUDIT_BUFFER_SIZE", 8192),
		AuditWorkerGoroutines: getInt("AUDIT_WORKERS", 4),
	}

	return cfg
}

func getString(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func mustGet(key string) string {
	value := getString(key, "")
	if value == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return value
}

func getInt(key string, fallback int) int {
	value := getString(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("invalid int value for %s: %v", key, err)
	}
	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := getString(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("invalid duration value for %s: %v", key, err)
	}
	return parsed
}
