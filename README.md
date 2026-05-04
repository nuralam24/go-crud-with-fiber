# High-Performance Go CRUD (Fiber + pgx + sqlc + PgBouncer)

Production-style Go API for low-latency CRUD with:
- **Authentication** (JWT)
- **Authorization** (RBAC: admin can `POST`, users/admin can `GET`)
- **Fiber** for fast HTTP handling
- **sqlc + pgx + PgBouncer** for type-safe high-performance data access
- **Goroutine worker pool** for async audit logging

## Architecture

```text
cmd/api                 -> composition root (bootstrap)
db/query                -> sqlc query source files
db/schema               -> sqlc schema source files
internal/config         -> typed env config
internal/domain         -> business entities/rules
internal/repository     -> postgres repository wrappers
internal/repository/postgres/sqlc -> generated typed query layer
internal/service        -> business use-cases + auth
internal/platform       -> db + async infra
internal/transport/http -> handlers, middleware, routes
migrations              -> SQL schema
```

## Why This Design For High Throughput

1. **Fiber + fasthttp** keeps allocation and latency low.
2. **sqlc + pgxpool** gives compile-time checked SQL and predictable runtime performance.
3. **Simple protocol mode** is enabled for PgBouncer transaction pooling compatibility.
4. **Async audit logging** avoids blocking the request path.
5. **Stateless JWT auth** avoids session-store bottlenecks.
6. **Layered architecture** keeps hot path clear and maintainable.

## API

### Login
`POST /v1/auth/login`

```json
{
  "email": "admin@gmail.com",
  "password": "12345"
}
```

Response:

```json
{
  "access_token": "<jwt>",
  "role": "admin"
}
```

### Create Item (Admin only)
`POST /v1/items`
Header: `Authorization: Bearer <token>`

```json
{
  "title": "item-1",
  "description": "demo"
}
```

### List Items (User/Admin)
`GET /v1/items?limit=20&offset=0`

### Get Item By ID (User/Admin)
`GET /v1/items/:id`

## Run

1. Copy env:
   - `cp .env.example .env`
2. Export env:
   - `set -a && source .env && set +a`
3. Apply migration:
   - `psql "$PG_DSN" -f migrations/001_init.sql`
4. Start API:
   - `go run ./cmd/api`
5. (Optional) Static checks before push/CI:
   - `make lint`
6. (Optional) Re-generate SQL layer after query changes:
   - `make sqlc`

## Load-Test Target Notes (1M requests/hour)

1M/hour is around **278 req/sec** average. Real systems should still design for 5-10x spikes.

Recommended for production:
- Run multiple API instances behind L4/L7 load balancer.
- Keep PgBouncer in transaction mode, right-size DB pool by CPU cores and query latency.
- Add Redis caching for hot GET paths if read-heavy.
- Add structured metrics (p95/p99 latency, DB saturation, auth failure rate).
- Add k6/Gatling test profile before release.
