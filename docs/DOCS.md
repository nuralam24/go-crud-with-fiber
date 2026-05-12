# Go CRUD (Fiber) — Project guide

This document is the **entry point** for understanding the repository. Deeper topics are split across linked files.

**Related docs**

- Architecture (layers, diagrams, scaling): [`ARCHITECTURE.md`](./ARCHITECTURE.md)  
- HTTP request → response (file-by-file): [`REQUEST_LIFECYCLE.md`](./REQUEST_LIFECYCLE.md)  
- Learning topics: [`LEARNING_TOPICS.md`](./LEARNING_TOPICS.md)  

---

## 1) What is this project?

A **Go + Fiber** REST API with PostgreSQL (**pgx** / **pgxpool**), **sqlc** for typed SQL, **JWT** access and refresh tokens, and **role-based** access for **admin** and **user**.

Highlights:

- Admin and user auth (login + refresh), user registration
- **Items**: list and get (admin + user), create (**admin only**) — no item update/delete routes in the current router
- **Brands**: list (admin + user), create (**admin only**) — no brand get-by-id / update / delete routes in the current router
- User profile: `GET` / `PATCH` `/api/v1/users/me` (**user** role)
- **Swagger UI** and embedded **OpenAPI** under `/api/docs`
- **Air** watch mode (`make watch`) for local development

Module: **`github.com/storex/go-crud`**

---

## 2) Tech stack

| Area | Technology |
|------|------------|
| Language | Go |
| HTTP | **Fiber v2** |
| DB | **pgx/v5** `pgxpool` (PgBouncer-friendly simple protocol where configured) |
| SQL | **sqlc** (generated queries under `internal/repository/postgres/sqlc`) |
| Auth | **JWT** (HS256), bcrypt; refresh tokens stored hashed |
| API docs | Embedded `swagger_openapi.yaml` + Swagger UI (CDN) |
| Hot reload | **Air** (`.air.toml`) |
| Logging | **`log/slog`** JSON to stdout |

---

## 3) Folder structure

### `cmd/`

- **`cmd/api/main.go`** — Loads config, connects to Postgres, starts the audit logger, builds the Fiber app, runs until shutdown.

### `internal/`

Go **`internal/`** convention: code here is not importable by other modules.

| Path | Role |
|------|------|
| **`internal/app/app.go`** | Composition root: Fiber app, global middleware, `RegisterRoutes`, `RegisterSwaggerRoutes`, `Run` (listen + graceful shutdown) |
| **`internal/config/`** | Environment → typed `Config` (timeouts, DSN, JWT, pool limits, audit workers) |
| **`internal/domain/`** | Entities and validation (e.g. items, brands, users) |
| **`internal/service/`** | Use cases: auth, users, items, brands |
| **`internal/repository/postgres/`** | Adapters: sqlc + mapping to domain types |
| **`internal/repository/postgres/sqlc/`** | **Generated** — do not edit by hand; run `make sqlc` |
| **`internal/platform/db/`** | Pool creation, ping / retry |
| **`internal/platform/async/`** | Buffered audit logger (`Start` / `Stop`, non-blocking `Publish`) |
| **`internal/transport/http/`** | Routes, handlers, middleware, **`error_handler.go`**, `swagger.go`, embedded OpenAPI |
| **`internal/transport/response/`** | Success JSON envelope (`response.Single`, etc.) |

### `migrations/`

- **`001_init.sql`** — DDL for the real database (applied with `make migrate` or your own process).

### `db/`

- **`db/schema/`** — Schema input for **sqlc** (compile-time typing; not applied to the DB at runtime).
- **`db/query/`** — SQL query definitions for sqlc.

### Repository root

| File | Purpose |
|------|---------|
| `.env` | Local secrets (gitignored) |
| `.env.example` | Example variables (no real secrets) |
| `Makefile` | `run`, `watch`, `build`, `test`, `fmt`, `lint`, `migrate`, `sqlc` (see section 11) |
| `.golangci.yml` | golangci-lint configuration |
| `sqlc.yaml` | sqlc project config |
| `.air.toml` | Air watch configuration |

The **Makefile** optionally **`include`s** `.env` and **`export`s** variables so tools like `make migrate` see `PG_DSN` when defined there.

---

## 4) Request flow (high level)

Example: `GET /api/v1/items`

1. Request hits the Fiber app.
2. Global middleware: **CORS** → **request ID** → **request logging**.
3. Route group middleware: **JWT authenticate** (protected routes), then **authorize** by role.
4. Handler parses query/body and calls the **service** layer.
5. Service calls the **repository** → **sqlc** → PostgreSQL.
6. Handler returns **`response.Single`** / JSON success shape.
7. Errors bubble to Fiber’s **`ErrorHandler`** (`error_handler.go`) for a consistent JSON error body.

---

## 5) API endpoints (current router)

Base path for versioned JSON API: **`/api/v1`**. Default server: **`http://127.0.0.1:8080`** (`HTTP_ADDR`, default `:8080`).

### Public / unversioned

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Service metadata + `swagger_ui`, `openapi_spec`, `api_base` |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness (DB ping) |

### Public under `/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/admin/login` | Admin login (email/password) |
| POST | `/api/v1/auth/admin/refresh` | Admin refresh token |
| POST | `/api/v1/users/register` | User registration |
| POST | `/api/v1/users/login` | User login |
| POST | `/api/v1/users/refresh` | User refresh token |

### Protected (requires `Authorization: Bearer <access_token>`)

| Method | Path | Roles |
|--------|------|-------|
| GET | `/api/v1/items` | admin, user |
| GET | `/api/v1/items/:id` | admin, user |
| POST | `/api/v1/items` | **admin only** |
| GET | `/api/v1/brands` | admin, user |
| POST | `/api/v1/brands` | **admin only** |
| GET | `/api/v1/users/me` | **user only** |
| PATCH | `/api/v1/users/me` | **user only** |

### Docs and redirects

| Path | Behavior |
|------|----------|
| `/api/docs` | Swagger UI |
| `/api/docs/openapi.yaml` | OpenAPI YAML |
| `/swagger`, `/docs` | Redirect to `/api/docs` |
| `/openapi.yaml` | Redirect to `/api/docs/openapi.yaml` |

For request/response schemas, use **Swagger** at **`/api/docs`**.

---

## 6) Auth and roles

Send **`Authorization: Bearer <access_token>`** on protected routes.

JWT claims include **role** (`admin` or `user`). Middleware **`Authenticate`** validates the token; **`Authorize`** enforces allowed roles per route.

Admin login response includes **`access_token`**, **`refresh_token`**, and **`role`** (see `handler/auth_handler.go`).

---

## 7) Database and sqlc

- Apply **`migrations/001_init.sql`** to your Postgres instance (`make migrate` if `PG_DSN` is set — see Makefile).
- After changing **`db/query`** or **`db/schema`**, run **`make sqlc`** to regenerate `internal/repository/postgres/sqlc/`.

Keep **migration DDL** and **sqlc schema** in sync when you change tables.

---

## 8) Error handling

Fiber **`ErrorHandler`** (`internal/transport/http/error_handler.go`) maps **`apierror.Error`**, **`fiber.Error`**, and other errors to JSON:

- `success: false`
- `error`: `code`, `http_status`, `message`, `data`
- `meta`: `request_id`, `timestamp`

5xx responses are logged with **`slog`**.

---

## 9) Logging and observability

Structured JSON logs via **`log/slog`**.

Per-request access logs (after the handler runs) include method, path, status, latency, client IP, and **`request_id`** (see `middleware/observability.go`).

---

## 10) How to run

1. Copy **`.env.example`** to **`.env`** and set at least **`PG_DSN`** and **`JWT_SECRET`** (see `internal/config/config.go` for all options).
2. Apply migrations:

```bash
make migrate
```

3. Run the API:

```bash
make run
```

4. Watch mode (rebuild on change):

```bash
make watch
```

5. Open docs: **`http://127.0.0.1:8080/api/docs`**

**CORS** in `internal/app/app.go` is configured for **`http://127.0.0.1:8080`** and **`http://localhost:8080`**. If you change **`HTTP_ADDR`**, update CORS origins to match your browser origin.

---

## 11) Makefile targets

| Target | Purpose |
|--------|---------|
| `make run` | Run `cmd/api` |
| `make watch` | Install Air if needed, run Air |
| `make build` | `go build ./...` |
| `make test` | `go test ./...` |
| `make fmt` | `go fmt ./...` |
| `make lint` | golangci-lint |
| `make migrate` | `psql "$(PG_DSN)" -f migrations/001_init.sql` |
| `make sqlc` | Regenerate sqlc code |

---

## 12) Why this layout helps

Patterns already in the repo:

- Layered architecture (transport → service → repository)
- Central configuration from environment
- JWT + RBAC middleware
- Structured logging and health endpoints
- OpenAPI contract + Swagger UI

Possible next steps (not required for local dev):

- Metrics (Prometheus), tracing (OpenTelemetry)
- Broader integration tests (e.g. testcontainers)
- Additional CRUD routes if you want parity with other services

---

## 13) Mental model (one line)

**Handlers parse HTTP → services enforce rules → repositories + sqlc talk to Postgres → success responses use `response` helpers; errors use Fiber’s centralized error handler.**
