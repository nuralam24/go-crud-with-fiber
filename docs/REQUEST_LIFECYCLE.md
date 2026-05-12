# Request → Response Lifecycle (Fiber)

This document explains **where an HTTP request starts**, **which files it passes through**, and **how the response is produced**.

---

## 1) Big picture

```text
Client (browser/curl)
    ↓ TCP
Fiber HTTP server (Listen)
    ↓
Global middleware (CORS → Request ID → Request log)
    ↓
Route match (router)
    ↓
Route-specific middleware (JWT Authenticate + Authorize)  [protected routes only]
    ↓
HTTP Handler (BodyParser → service)
    ↓
Service (business rules → repository)
    ↓
Repository (sqlc → PostgreSQL)
    ↓
Success: response helpers (e.g. response.Single) OR Error → Fiber ErrorHandler
```

---

## 2) Server start (before any request)

These steps run **once** when you run `make run` or `make watch`.

| Step | File | What happens |
|------|------|----------------|
| 1 | `cmd/api/main.go` | `config.Load()` |
| 2 | `internal/config/config.go` | Build `Config` from environment |
| 3 | `cmd/api/main.go` | `db.NewPostgresPool(...)` |
| 4 | `internal/platform/db/postgres.go` | Create pool, ping/retry |
| 5 | `cmd/api/main.go` | `async.NewAuditLogger` + `Start()`; `defer auditLogger.Stop()` |
| 6 | `internal/platform/async/audit_logger.go` | Background workers; non-blocking `Publish` |
| 7 | `cmd/api/main.go` | `app.NewServer(...)` |
| 8 | `internal/app/app.go` | `fiber.New(...)` with **`ErrorHandler`**, timeouts; `Use` CORS, identity, logger; `RegisterRoutes`; `RegisterSwaggerRoutes` |
| 9 | `internal/app/app.go` | `Run`: goroutine waits on `ctx.Done()` then **`ShutdownWithContext`**; **`app.Listen(cfg.HTTPAddr)`** |

The request lifecycle begins **after step 9**, when a client connects and sends HTTP.

---

## 3) When a request arrives (inside Fiber)

Fiber builds **`fiber.Ctx`** per request (headers, body, path, response helpers).

**Global middleware** order (current code): `internal/app/app.go`

1. **CORS** — `github.com/gofiber/fiber/v2/middleware/cors`
2. **`RequestIdentity`** — `internal/transport/http/middleware/observability.go`  
   - Set or generate **`X-Request-ID`**  
   - Store in **`c.Locals("request_id", ...)`**
3. **`RequestLogger`** — same file  
   - Measure latency around **`c.Next()`**  
   - Log method, path, status, latency, IP

Then the **matched route** runs (including route-level middleware).

---

## 4) Routing (URL → handler)

File: **`internal/transport/http/router.go`**

| Path | Handler / middleware |
|------|------------------------|
| `GET /` | Inline JSON (`service`, `swagger_ui`, `openapi_spec`, `api_base`) |
| `GET /healthz` | `HealthHandler.Liveness` |
| `GET /readyz` | `HealthHandler.Readiness` |
| `POST /api/v1/auth/admin/login` | `AuthHandler.AdminLogin` |
| `POST /api/v1/auth/admin/refresh` | `AuthHandler.AdminRefresh` |
| `POST /api/v1/users/register` | `UserHandler.Register` |
| `POST /api/v1/users/login` | `UserHandler.Login` |
| `POST /api/v1/users/refresh` | `UserHandler.Refresh` |
| Under `v1`, group **`protected`** | **`middleware.Authenticate`**, then per-route **`Authorize(...)`** |

Protected routes today include: **items** (list/get/create), **brands** (list/create), **users/me** (get/patch). See `router.go` for the exact matrix.

Swagger:

- **`internal/transport/http/swagger.go`** — `/api/docs`, `/api/docs/openapi.yaml`, redirects  
- **`swagger_openapi.yaml`** — embedded spec (`//go:embed` pattern in `swagger.go`)

---

## 5) Example A — `POST /api/v1/auth/admin/login` (no JWT on route)

1. `router.go` → **`AuthHandler.AdminLogin`**
2. **`handler/auth_handler.go`** — `BodyParser` into `loginRequest`, trim fields
3. Email must match configured admin identity; **`authService.Login`** issues tokens (DB-backed auth flow in service/repo)
4. Success: **`response.Single`** with `access_token`, `refresh_token`, `role`  
5. Errors: return **`apierror.New(...)`** (or wrapped errors) → **`ErrorHandler`**

---

## 6) Example B — `GET /api/v1/items` (protected + RBAC)

1. **`router.go`** — `protected` group uses **`middleware.Authenticate(authService)`**; route adds **`Authorize(RoleAdmin, RoleUser)`**
2. **`middleware/auth.go` — `Authenticate`** — read **`Authorization: Bearer`**, **`authService.ParseToken`**, set claims in **`c.Locals`**
3. **`Authorize`** — deny with **403** if role not allowed
4. **`handler/item_handler.go` — `List`** — query params **`limit`** (default 20) and **`offset`** (default 0), then **`itemService.List(c.Context(), ...)`**
5. **`service/item_service.go`** — clamp pagination, call repo
6. **`repository/postgres/item_repository.go`** + **`sqlc`** — SQL via **`pgxpool`**
7. Handler returns **`c.JSON`** / **`response`** helpers as implemented
8. **`RequestLogger`** logs status and latency after **`c.Next()`** returns

---

## 7) Example C — `POST /api/v1/items` (admin only)

Same as Example B until **`Authorize`**:

- **`Authorize(RoleAdmin)`** — **user** role gets **403** before the handler runs.

Then **`ItemHandler.Create`** → **`ItemService.Create`** → repository **`INSERT`**; audit may **`Publish`** asynchronously.

---

## 8) What happens on error

Handlers and middleware typically return **`error`**:

- **`apierror.Error`** — mapped in **`ErrorHandler`** to `success: false` JSON with code, HTTP status, message, and **`meta.request_id`**
- **`fiber.NewError(status, msg)`** — same pipeline with status-derived codes
- **5xx** — logged at error level with structured fields

Implementation: **`internal/transport/http/error_handler.go`**.

---

## 9) SQL layer (`sqlc`)

| Layer | Path | Role |
|-------|------|------|
| Author SQL | `db/query/*.sql` | Source of truth for queries |
| Author schema | `db/schema/*.sql` | Types / compile checks for sqlc |
| Generated | `internal/repository/postgres/sqlc/*.go` | Typed methods (**do not edit**) |
| Adapters | `internal/repository/postgres/*_repository.go` | Map sqlc models ↔ domain |

After SQL changes: **`make sqlc`**. Apply DDL with **`make migrate`** or your own migration tool.

---

## 10) Trace checklist

1. `cmd/api/main.go`  
2. `internal/app/app.go` — middleware order, `ErrorHandler`, `Run`  
3. `internal/transport/http/router.go` — exact routes  
4. `internal/transport/http/middleware/auth.go`  
5. `internal/transport/http/handler/*.go`  
6. `internal/service/*.go`  
7. `internal/repository/postgres/` + `sqlc/`  
8. `internal/transport/http/error_handler.go`  

---

## Related

- [DOCS.md](./DOCS.md)  
- [ARCHITECTURE.md](./ARCHITECTURE.md)  
