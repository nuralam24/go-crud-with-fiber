# Request → Response Lifecycle

এই ডকুমেন্টে বোঝানো হয়েছে: **একটা HTTP request এসে কোথা থেকে শুরু হয়, কোন কোন ফাইল দিয়ে গিয়ে response ফেরত যায়**।

---

## 1) Big picture (এক নজরে)

```text
Client (browser/curl)
    ↓ TCP
Fiber HTTP server (Listen)
    ↓
Global middleware chain (CORS → Request ID → Request log)
    ↓
Route match (router)
    ↓
Route-specific middleware (JWT auth / RBAC)  [শুধু protected routes]
    ↓
HTTP Handler (parse input → call service)
    ↓
Service (business rules → call repository)
    ↓
Repository (sqlc queries → PostgreSQL)
    ↓
Response JSON (handler) অথবা Error → Fiber ErrorHandler
```

---

## 2) Server start (request আসার আগে)

এগুলো **একবার** চালে যখন তুমি `make run` / `make watch` দাও।

| Step | File | কী হয় |
|------|------|--------|
| 1 | `cmd/api/main.go` | `main()` → `config.Load()` |
| 2 | `internal/config/config.go` | env থেকে `Config` struct |
| 3 | `cmd/api/main.go` | `db.NewPostgresPool(...)` |
| 4 | `internal/platform/db/postgres.go` | `pgxpool` তৈরি + ping/retry |
| 5 | `cmd/api/main.go` | `async.NewAuditLogger` + `Start` |
| 6 | `internal/platform/async/audit_logger.go` | background goroutine workers |
| 7 | `cmd/api/main.go` | `app.NewServer(...)` |
| 8 | `internal/app/app.go` | Fiber app, middleware, routes wire-up |
| 9 | `internal/app/app.go` | `app.Run(...)` → `Listen` |

Request lifecycle শুরু হয় যখন **step 9** এর পর client TCP connection খুলে HTTP পাঠায়।

---

## 3) একটা request এলে (Fiber এর ভিতরে)

Fiber প্রথমে `*fiber.Ctx` বানায় — এটাই এক request এর context (headers, body, path params, response writer)।

তারপর **global middleware** গুলো **উপর থেকে নিচে** চলে (`app.Use` order অনুযায়ী)।

Order (বর্তমান কোড): `internal/app/app.go`

1. **CORS** — `github.com/gofiber/fiber/v2/middleware/cors` (third-party, `app.go` এ configure)
2. **`RequestIdentity`** — `internal/transport/http/middleware/observability.go`
   - `X-Request-ID` set / generate
   - `c.Locals("request_id", ...)`
3. **`RequestLogger`** — same file
   - `c.Next()` এর আগে-পরে time measure
   - response status + latency log

তারপর **route handler** চলে।

---

## 4) Routing (কোন URL কোথায় map)

File: `internal/transport/http/router.go`

| Path | Handler / middleware |
|------|------------------------|
| `GET /` | inline JSON (service links) |
| `GET /healthz` | `handler.HealthHandler.Liveness` |
| `GET /readyz` | `handler.HealthHandler.Readiness` |
| `POST /v1/auth/login` | `handler.AuthHandler.Login` |
| `GET/POST ...` under `/v1` protected group | `middleware.Authenticate` তারপর per-route `middleware.Authorize` |

Swagger:

- `internal/transport/http/swagger.go` — `RegisterSwaggerRoutes` (`/swagger`, `/openapi.yaml`)
- `internal/transport/http/swagger_openapi.yaml` — OpenAPI spec (embed)

---

## 5) Example A — `POST /v1/auth/login` (no JWT middleware)

Flow:

1. `router.go` → `authHandler.Login`
2. `internal/transport/http/handler/auth_handler.go`
   - JSON body parse
   - `authService.Login(...)`
3. `internal/service/auth_service.go`
   - credential check → JWT sign
4. Response: JSON `{ access_token, role }`

এখানে **DB hit হয় না** (বর্তমানে env-based demo users)।

---

## 6) Example B — `GET /v1/items` (protected + RBAC)

Flow (ফাইল অনুসারে):

1. **`router.go`**
   - `v1` group এর ভিতরে `protected` group: `middleware.Authenticate(authService)`

2. **`internal/transport/http/middleware/auth.go` — `Authenticate`**
   - `Authorization: Bearer ...` read
   - `authService.ParseToken` (`internal/service/auth_service.go`)
   - success হলে `c.Locals(claimsContextKey, claims)`

3. **`router.go` — route middleware `Authorize(...)`**
   - same file `middleware/auth.go` — role allowlist check

4. **`internal/transport/http/handler/item_handler.go` — `List`**
   - query params parse (`limit`, `offset`)
   - `itemService.List(c.Context(), ...)`

5. **`internal/service/item_service.go` — `List`**
   - limit/offset clamp
   - `repo.List(...)`

6. **`internal/repository/postgres/item_repository.go` — `List`**
   - `queries.ListItems(...)` — এটা sqlc generated call

7. **`internal/repository/postgres/sqlc/items.sql.go` (generated)**
   - actual SQL `SELECT ... LIMIT ... OFFSET ...`
   - PostgreSQL এ query চলে (`pgxpool` দিয়ে)

8. **Response path উল্টো দিকে**
   - sqlc `[]Item` → `toDomainItem` → `[]domain.Item`
   - handler `c.JSON(...)`

9. **`RequestLogger` middleware return**
   - access log line print (status + latency)

---

## 7) Example C — `POST /v1/items` (admin only)

Same as Example B until authorize:

- `Authorize(service.RoleAdmin)` — user হলে এখানেই **403** এবং handler পর্যন্ত যাবে না।

Handler:

- `internal/transport/http/handler/item_handler.go` — `Create`
- `internal/service/item_service.go` — `Create` (domain validation + `repo.Create`)
- `internal/domain/item.go` — validation rules
- `internal/repository/postgres/item_repository.go` + `sqlc` — `INSERT`
- `internal/platform/async/audit_logger.go` — `Publish` (non-blocking audit event)

---

## 8) Error হলে কী হয়

যেকোনো handler/middleware `return fiber.NewError(...)` বা raw `error` return করলে:

1. Fiber **`ErrorHandler`** call করে
2. File: `internal/transport/http/error_handler.go` — `NewErrorHandler` closure
   - status + message decide
   - 5xx হলে `slog.Error(...)` structured log
   - client এ JSON: `{ "error": "...", "request_id": "..." }`

Middleware chain এর ভিতরে error হলেও একই pattern — Fiber error handler পর্যন্ত যায়।

---

## 9) SQL layer (sqlc) কোথায় বসে

| Layer | Path | Role |
|-------|------|------|
| Human-written SQL | `db/query/*.sql` | queries (source of truth) |
| Human-written schema (sqlc) | `db/schema/*.sql` | types + compile checks |
| Generated Go | `internal/repository/postgres/sqlc/*.go` | typed query methods (**do not edit**) |
| Adapter | `internal/repository/postgres/item_repository.go` | domain mapping + pool injection |

SQL change করলে: `make sqlc` → generated code update।

Runtime DB schema apply: `make migrate` (`migrations/001_init.sql`)।

---

## 10) দ্রুত চেকলিস্ট (নিজে trace করতে চাইলে)

1. `cmd/api/main.go` — process entry
2. `internal/app/app.go` — middleware order + `RegisterRoutes`
3. `internal/transport/http/router.go` — exact URL → handler mapping
4. `internal/transport/http/middleware/auth.go` — JWT + RBAC
5. `internal/transport/http/handler/*.go` — HTTP ↔ service boundary
6. `internal/service/*.go` — business logic
7. `internal/repository/postgres/item_repository.go` + `sqlc/` — DB
8. `internal/transport/http/error_handler.go` — failure response shape

---

## Related

- Full project tour: `docs/DOCS.md`
