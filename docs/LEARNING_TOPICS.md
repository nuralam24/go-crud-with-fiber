# What you can learn from this project (Fiber)

This document lists **topics the codebase actually uses**. Understanding them well is enough to build similar HTTP APIs in Go.

Each section includes: **what to learn**, **where it appears here**, **practice ideas**.

---

## Topic map (quick index)

| # | Topic | Main files / folders |
|---|--------|------------------------|
| 1 | Go modules, `package`, `internal/` | `go.mod`, `cmd/`, `internal/` |
| 2 | `main`, bootstrap, composition | `cmd/api/main.go`, `internal/app/app.go` |
| 3 | `defer` | `cmd/api/main.go` |
| 4 | `context.Context` | `main`, `app.Run`, handlers, DB |
| 5 | Signals + graceful shutdown | `cmd/api/main.go`, `internal/app/app.go` |
| 6 | Structs, `New*` constructors | All layers |
| 7 | Small interfaces + dependency direction | `internal/service/*_service.go` |
| 8 | Errors: `apierror`, `fiber.Error`, `ErrorHandler` | `handler/`, `error_handler.go`, `service/` |
| 9 | JSON + Fiber handlers | `internal/transport/http/handler/` |
| 10 | Middleware chain | `internal/app/app.go`, `middleware/` |
| 11 | JWT authentication | `internal/service/auth_service.go`, `middleware/auth.go` |
| 12 | RBAC | `middleware/auth.go` |
| 13 | PostgreSQL + `pgxpool` | `internal/platform/db/postgres.go` |
| 14 | `sqlc` workflow | `db/`, `sqlc.yaml`, `internal/repository/postgres/` |
| 15 | Domain validation | `internal/domain/` |
| 16 | Goroutines, channels, `select`, `WaitGroup` | `internal/platform/async/audit_logger.go` |
| 17 | Structured logging (`slog`) | `cmd/api/main.go`, `middleware/`, `error_handler.go` |
| 18 | Environment config | `internal/config/config.go`, `.env` |
| 19 | Migrations vs sqlc schema | `migrations/`, `db/schema/` |
| 20 | Dev tooling (Make, Air) | `Makefile`, `.air.toml` |

---

## 1) Go modules, packages, `internal/`

**What to learn**

- `go.mod` declares **`module github.com/storex/go-crud`** — base import path.
- `package main` for executables only.
- Code under **`internal/`** cannot be imported by other modules — encapsulation.

**In this repo**

- `go.mod`, `cmd/api/main.go`
- Application code under `internal/`.

**Practice**

- Add `internal/demo` and call it from `main` to see how imports resolve.

---

## 2) Small `main`, composition in `internal/app`

**What to learn**

- `main`: config → DB pool → audit logger → **`NewServer`** → **`Run`**.
- **`app.NewServer`** is the composition root (Fiber, middleware, handlers, routes).

**In this repo**

- `cmd/api/main.go` — short bootstrap.
- `internal/app/app.go` — Fiber app creation and wiring.

**Practice**

- Decide where a new dependency (for example a cache client) should be constructed — `main` vs `app.go`.

---

## 3) `defer`

**What to learn**

- `defer` runs when the surrounding function returns (LIFO). Use for cleanup: `Close()`, `Stop()`, cancel funcs.

**In this repo**

- `cmd/api/main.go`: `defer stop()`, `defer pool.Close()`, `defer auditLogger.Stop()`.

**Practice**

- Reason about defer order when multiple defers run on exit.

---

## 4) `context.Context`

**What to learn**

- Cancellation, deadlines, passing `ctx` into I/O (DB, outbound HTTP).

**In this repo**

- `signal.NotifyContext` in `main`.
- `app.Run` waits on `ctx.Done()` then **`app.ShutdownWithContext`**.
- Handlers pass **`c.Context()`** (or `c.UserContext()` if you standardize on that) into services.

**Practice**

- Trace one DB call and confirm `ctx` is forwarded to `pgx`.

---

## 5) Graceful shutdown

**What to learn**

- On SIGINT/SIGTERM, stop accepting new work and drain in-flight requests where possible.

**In this repo**

- Goroutine in `Run` calls **`ShutdownWithContext`** with `cfg.ShutdownTimeout`.

**Practice**

- Hit Ctrl+C while a slow request runs and observe logs.

---

## 6) Fiber basics

**What to learn**

- **`fiber.New(fiber.Config{...})`** — global **`ErrorHandler`**, timeouts, `AppName`.
- **`app.Use` / `app.Get` / `app.Post`** — routing and middleware.
- **`c.BodyParser`**, **`c.JSON`**, **`c.Params`**, **`c.Query`**, **`c.Locals`**.

**In this repo**

- `internal/app/app.go`, `internal/transport/http/router.go`, `handler/`.

**Practice**

- Add a trivial `GET /version` route and return JSON.

---

## 7) Interfaces in the service layer

**What to learn**

- Services depend on small repository **interfaces** for testability.

**In this repo**

- See `internal/service/item_service.go` (and similar) for `ItemRepository`-style interfaces.

**Practice**

- Confirm services do not import concrete `postgres` packages directly.

---

## 8) Centralized errors (`apierror` + `ErrorHandler`)

**What to learn**

- Return **`apierror.New(status, code, message)`** from handlers for predictable client errors.
- Fiber **`ErrorHandler`** maps errors to one JSON shape.

**In this repo**

- `internal/transport/apierror/`, `internal/transport/http/error_handler.go`.

**Practice**

- Add a new domain error and map it through `apierror` in a handler.

---

## 9) JWT + RBAC

**What to learn**

- HS256 JWTs, **`Authorization: Bearer`**, role claims, refresh token flows as implemented.

**In this repo**

- `internal/service/auth_service.go`, `middleware/auth.go`.

---

## 10) PostgreSQL + `sqlc`

**What to learn**

- Connection pooling, **`make sqlc`**, adapters between sqlc rows and domain types.

**In this repo**

- `internal/platform/db/postgres.go`, `db/query`, `internal/repository/postgres/sqlc/`.

**Practice**

- Change a query, run `make sqlc`, fix compile errors in repositories.

---

## 11) Audit workers and `slog`

**What to learn**

- Background workers with channels; non-blocking send under pressure; structured logs.

**In this repo**

- `internal/platform/async/audit_logger.go`, `middleware/observability.go`.

---

## 12) Migrations vs sqlc schema

**What to learn**

- **Migrations** change the real database.
- **`db/schema`** is for sqlc only — keep them aligned when you evolve tables.

**In this repo**

- `migrations/001_init.sql`, `db/schema/*.sql`.

---

## 13) Makefile and Air

**What to learn**

- One-command workflows: **`make run`**, **`make migrate`**, **`make sqlc`**, **`make watch`**.

**In this repo**

- `Makefile` (note optional **`include .env`** + **`export`** for subprocesses), `.air.toml`.

---

## Minimum checklist for your own API

1. `main` + composition root (`internal/app` or equivalent)  
2. `context` + graceful shutdown  
3. `defer` for pool / background workers  
4. Router + at least one handler  
5. Service + repository interfaces  
6. DB pool + migrations  
7. (Recommended) `sqlc`  
8. Central error JSON + structured logs  
9. `/healthz` and `/readyz`  
10. JWT + roles if you have multiple actors  

---

## Going deeper (outside this repo)

- Integration tests with **`httptest`** or **testcontainers**  
- Metrics / tracing (Prometheus, OpenTelemetry)  
- Rate limiting, idempotency keys  

---

## Related docs

- [DOCS.md](./DOCS.md)  
- [REQUEST_LIFECYCLE.md](./REQUEST_LIFECYCLE.md)  
- [ARCHITECTURE.md](./ARCHITECTURE.md)  
