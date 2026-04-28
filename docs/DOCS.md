# Go CRUD Project Full Guide (Beginner Friendly)

এই ডকুমেন্টটি এমনভাবে লেখা হয়েছে যেন Go সম্পর্কে আগে থেকে ধারণা না থাকলেও পুরো প্রজেক্টটা বোঝা যায়।

**Related docs**
- আর্কিটেকচার (লেয়ার, ডায়াগ্রাম, স্কেলিং নোট): [`ARCHITECTURE.md`](./ARCHITECTURE.md)
- HTTP request → response (ফাইল বাই ফাইল): [`REQUEST_LIFECYCLE.md`](./REQUEST_LIFECYCLE.md)
- শেখার টপিক (এই প্রজেক্ট থেকে নিজের প্রজেক্ট): [`LEARNING_TOPICS.md`](./LEARNING_TOPICS.md)

---

## 1) Project ta ki?

এটা একটি **high-performance CRUD API** project, যেটা Go language + Fiber framework দিয়ে বানানো।

Main features:
- Login করে JWT token নেওয়া
- Role-based access:
  - `admin` -> `POST /v1/items` করতে পারে
  - `admin` + `user` -> `GET /v1/items`, `GET /v1/items/:id` করতে পারে
- PostgreSQL (Neon compatible) database
- Swagger UI দিয়ে API test
- Watch mode (`make watch`) দিয়ে auto-reload development

---

## 2) Tech stack

- **Language:** Go
- **HTTP framework:** Fiber
- **DB driver/pool:** pgx + pgxpool
- **SQL code generation:** sqlc (typed query layer)
- **DB proxy compatibility:** PgBouncer-friendly config
- **Auth:** JWT
- **API docs:** Swagger (OpenAPI)
- **Hot reload:** Air
- **Logging:** structured JSON log (`log/slog`)

---

## 3) Folder structure (কোন folder কেন)

### `cmd/`
Application entrypoint রাখার standard জায়গা।

- `cmd/api/main.go`
  - app start করে
  - config load করে
  - DB connect করে
  - server run করে

### `internal/`
Business/application related সব core code এখানে। `internal` মানে project-এর বাইরে থেকে import করা যাবে না (Go best practice)।

- `internal/app/`
  - `app.go`
  - Application composition root
  - সব dependency এক জায়গায় wire-up করা হয় (repo, service, handler, middleware)

- `internal/config/`
  - `config.go`
  - `.env`/environment variable থেকে config পড়ে typed struct বানায়
  - timeouts, DB config, JWT secret, credentials ইত্যাদি এখান থেকে আসে

- `internal/domain/`
  - `item.go`
  - business entity (`Item`) এবং validation rules

- `internal/repository/postgres/`
  - `item_repository.go`
  - repository wrapper layer
  - service/domain friendly format-এ data map করে

- `internal/repository/postgres/sqlc/`
  - sqlc generated Go files (auto-generated, hand-edit না করা)
  - compile-time type-safe query methods
  - low-level DB read/write implementation

- `internal/service/`
  - `auth_service.go`: login, JWT generate/parse, role claims
  - `item_service.go`: item create/list/get business logic
  - service layer repository call করে এবং business rule enforce করে

- `internal/platform/`
  - cross-cutting infra concern
  - `db/postgres.go`: pgxpool init + ping/retry logic
  - `async/audit_logger.go`: goroutine-based async audit log

- `internal/transport/http/`
  - HTTP/API layer
  - `router.go`: route map করে
  - `error_handler.go`: central error handler
  - `swagger.go`: swagger page serve করে
  - `swagger_openapi.yaml`: API docs schema
  - `handler/`: request->service call
  - `middleware/`: auth, RBAC, request log, request id etc.

### `migrations/`
- `001_init.sql`
- DB table/index create script

### `db/`
- `db/schema/`
  - sqlc-এর জন্য schema source
- `db/query/`
  - sqlc query files (`-- name: ...` style)

### Root files

- `.env` -> local runtime secret/config (gitignore করা)
- `.env.example` -> sample config (real secret ছাড়া)
- `.gitignore` -> কোন file git-এ যাবে না
- `Makefile` -> shortcut commands (`run`, `watch`, `build`, `test`, `migrate`, `sqlc`)
- `sqlc.yaml` -> sqlc configuration
- `.air.toml` -> watch mode config
- `README.md` -> quick overview
- `DOCS.md` -> full guide (এই file)

---

## 4) Request flow (high-level)

উদাহরণ: `GET /v1/items`

1. Request আসে Fiber app-এ
2. Middleware run হয়:
   - request id
   - request logging
   - auth token validate
   - role authorize
3. HTTP handler request parse করে service call করে
4. Service repository call করে
5. Repository sqlc generated query call করে DB query run করে
6. Data response হিসেবে client এ ফিরে যায়
7. Error হলে centralized error handler JSON error return করে

---

## 5) API endpoints

### Auth
- `POST /v1/auth/login`
  - body: `{ "email": "...", "password": "..." }`
  - response: `{ "access_token": "...", "role": "admin|user" }`

### Item
- `POST /v1/items` (admin only)
- `GET /v1/items` (admin/user)
- `GET /v1/items/:id` (admin/user)

### Health
- `GET /healthz` -> process live আছে কিনা
- `GET /readyz` -> DB reachable কিনা

### Docs
- `GET /swagger`
- `GET /openapi.yaml`

---

## 6) Auth & Role details

Login successful হলে JWT token generate হয়।  
এই token `Authorization: Bearer <token>` header এ পাঠাতে হয়।

Role claim token-এর ভেতরে থাকে:
- `admin`
- `user`

RBAC middleware role check করে route access allow/deny করে।

---

## 7) Database details

Table: `items`
- `id` (UUID, PK)
- `title`
- `description`
- `created_at`
- `updated_at`

Index:
- `created_at DESC` index list endpoint কে efficient করে।

sqlc flow:
1. `db/schema` + `db/query` update করো
2. `make sqlc` run করো
3. generated code `internal/repository/postgres/sqlc` এ update হবে

---

## 8) Error handling

Central error handler:
- সব unhandled error কে consistent JSON format-এ return করে
- internal error হলে sanitized message দেয়
- `request_id` response-এ attach হয় (debug trace)
- console log-এ structured error print হয়

---

## 9) Logging & observability

Project structured JSON log ব্যবহার করে।

প্রতি request এ log fields থাকে:
- method
- path
- status
- latency
- ip
- request_id

এই approach production debugging/monitoring এ অনেক helpful।

---

## 10) Run guide (step-by-step)

1. env configure:
   - `.env` file update
   - `PG_DSN`, `JWT_SECRET` set

2. migration:
```bash
make migrate
```

3. normal run:
```bash
make run
```

4. watch mode:
```bash
make watch
```

5. docs open:
- `http://127.0.0.1:8080/swagger`

---

## 11) Make commands

- `make run` -> app start
- `make watch` -> auto-reload dev mode
- `make build` -> build check
- `make test` -> test run
- `make fmt` -> go fmt
- `make migrate` -> migration apply
- `make sqlc` -> SQL থেকে typed code generate

---

## 12) Professional notes (why this is good)

এই project-এ already কিছু strong professional pattern আছে:
- layered architecture
- centralized config
- clean repository/service split
- middleware based auth/authorization
- structured logging
- readiness/liveness health endpoints
- swagger contract

আরও enterprise-level করতে চাইলে next:
- DB-backed user table + hashed password
- refresh token flow
- metrics (Prometheus)
- tracing (OpenTelemetry)
- integration tests (testcontainers)

---

## 13) Quick mental model (এক লাইনে)

**Handler request নেয় -> Service business rule চালায় -> Repository DB query করে -> Response ফেরত দেয়.**

