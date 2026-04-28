# এই প্রজেক্ট থেকে যা শিখলে নিজের প্রজেক্ট করতে পারবে

এই ডকের লক্ষ্য: **Go এর হাজার হাজার টপিক না**, বরং **এই কোডবেজে যা যা ব্যবহার হয়েছে** — সেগুলো ভালোভাবে বুঝলে তুমি একই ধরনের API/সার্ভিস নিজে বানাতে পারবে।

প্রতিটি সেকশনে আছে: **কী শিখবে**, **এই প্রজেক্টে কোথায় দেখা যায়**, **নিজে প্র্যাকটিস করলে কী করবে**।

---

## টপিক ম্যাপ (দ্রুত তালিকা)

| # | টপিক | প্রধান ফাইল / ফোল্ডার |
|---|--------|------------------------|
| 1 | Go মডিউল, `package`, `internal/` | `go.mod`, `cmd/`, `internal/` |
| 2 | `main`, বুটস্ট্র্যাপ, কম্পোজিশন | `cmd/api/main.go`, `internal/app/app.go` |
| 3 | `defer` | `cmd/api/main.go` |
| 4 | `context.Context` | `cmd/api/main.go`, `internal/app/app.go`, handlers, DB |
| 5 | সিগনাল + গ্রেসফুল শাটডাউন | `cmd/api/main.go`, `internal/app/app.go` |
| 6 | স্ট্রাক্ট, কনস্ট্রাক্টর `New*` | সব লেয়ার |
| 7 | ইন্টারফেস + ডিপেন্ডেন্সি দিক নির্দেশ | `internal/service/item_service.go` |
| 8 | এরর হ্যান্ডলিং | `internal/service/`, `internal/transport/http/error_handler.go` |
| 9 | JSON + HTTP (Fiber) | `internal/transport/http/handler/` |
| 10 | মিডলওয়্যার চেইন | `internal/app/app.go`, `internal/transport/http/middleware/` |
| 11 | JWT অথেন্টিকেশন | `internal/service/auth_service.go`, `middleware/auth.go` |
| 12 | RBAC (রোল চেক) | `internal/transport/http/middleware/auth.go` |
| 13 | PostgreSQL + `pgxpool` | `internal/platform/db/postgres.go` |
| 14 | `sqlc` ওয়ার্কফ্লো | `db/`, `sqlc.yaml`, `internal/repository/postgres/` |
| 15 | ডোমেইন ভ্যালিডেশন | `internal/domain/item.go` |
| 16 | গোরুটিন, চ্যানেল, `select`, `sync.WaitGroup` | `internal/platform/async/audit_logger.go` |
| 17 | স্ট্রাকচার্ড লগ (`slog`) | `cmd/api/main.go`, `internal/transport/http/` |
| 18 | এনভায়রনমেন্ট কনফিগ | `internal/config/config.go`, `.env` |
| 19 | মাইগ্রেশন vs `sqlc` স্কিমা | `migrations/`, `db/schema/` |
| 20 | ডেভ টুলিং (Make, Air) | `Makefile`, `.air.toml` |

নিচে **বিস্তারিত**।

---

## 1) Go মডিউল, প্যাকেজ, `internal/`

**কী শিখবে**
- `go.mod` এ মডিউল path (`module github.com/storex/go-crud`) — import path এর ভিত্তি।
- `package main` শুধমাত্র executable এর জন্য।
- `internal/` — Go convention: বাইরের প্রজেক্ট থেকে `internal/...` import করা যায় না; অ্যাপের কোড এনক্যাপসুলেট থাকে।

**এই প্রজেক্টে**
- `go.mod`, `cmd/api/main.go` (`package main`)
- বাকি সব `package app`, `package service` ইত্যাদি `internal/` এর নিচে।

**প্র্যাকটিস**
- নতুন প্যাকেজ `internal/foo` বানিয়ে `main` থেকে কল করো; import path ঠিক আছে কিনা দেখো।

---

## 2) `main` ছোট রাখা, `internal/app` এ কম্পোজিশন

**কী শিখবে**
- `main` শুধু: config → infra (DB) → background workers → `NewServer` → `Run`।
- “সব কিছু এক জায়গায়” না করে **composition root** (`app.NewServer`) এ wire-up।

**এই প্রজেক্টে**
- `cmd/api/main.go` — সংক্ষিপ্ত বুটস্ট্র্যাপ।
- `internal/app/app.go` — Fiber, middleware, handler, service তৈরি ও রুট রেজিস্টার।

**প্র্যাকটিস**
- নতুন ডিপেন্ডেন্সি (যেমন ক্যাশ ক্লায়েন্ট) যোগ করলে কোথায় `New*` কল করবে — শুধু `app.go` vs `main.go` ঠিক করো।

---

## 3) `defer`

**কী শিখবে**
- `defer` ফাংশন শেষ হওয়ার আগে চলে (LIFO)। রিসোর্স ক্লিনআপ: `Close()`, `stop()`, ইত্যাদি।

**এই প্রজেক্টে**
- `cmd/api/main.go`: `defer stop()`, `defer pool.Close()`, `defer auditLogger.Stop()`  
  অর্থাৎ `main` শেষ/প্যানিক হলেও ক্লিনআপের চেষ্টা।

**প্র্যাকটিস**
- `defer` order মেনে চিন্তা করো: কোনটা আগে execute হবে।

---

## 4) `context.Context`

**কী শিখবে**
- ক্যানসেলেশন, টাইমআউট, রিকোয়েস্ট-স্কোপড ডেডলাইন।
- DB/API কলে `ctx` পাস করা — ক্লায়েন্ট চলে গেলে কাজ বন্ধ করতে পারে।

**এই প্রজেক্টে**
- `signal.NotifyContext(...)` — SIGINT/SIGTERM এ `ctx` cancel।
- `app.Run` → `ShutdownWithContext(shutdownCtx)`।
- `handler`: `c.Context()` পাস করে service/repo তে।
- `internal/platform/db/postgres.go` — ping এ `context.WithTimeout`।
- `audit_logger.go` — worker গুলো `ctx.Done()` দেখে বন্ধ।

**প্র্যাকটিস**
- নতুন DB কলে সবসময় `ctx` পাস করো; `context.Background()` শুধু যেখানে উপযুক্ত।

---

## 5) সিগনাল + গ্রেসফুল শাটডাউন

**কী শিখবে**
- প্রসেস kill করার আগে লিসেনার বন্ধ, কানেকশন ড্রেন — প্রোডাকশনে গুরুত্বপূর্ণ।

**এই প্রজেক্টে**
- `main`: `signal.NotifyContext` + `app.Run` এর ভিতরে goroutine যা `ctx.Done()` এর পর `fiber.ShutdownWithContext` চালায়।

**প্র্যাকটিস**
- Ctrl+C দিয়ে দেখো লগ/বিহেভিয়ার; `SHUTDOWN_TIMEOUT` env বুঝো।

---

## 6) স্ট্রাক্ট ও `New*` কনভেনশন

**কী শিখবে**
- `type X struct { ... }`, `func NewX(...) *X` — explicit constructor, zero value এ নির্ভর না করা।

**এই প্রজেক্টে**
- `NewItemRepository`, `NewItemService`, `NewAuthHandler`, `NewServer` ইত্যাদি।

**প্র্যাকটিস**
- নতুন সার্ভিস বানালে `NewFooService(deps...)` প্যাটার্ন ফলো করো।

---

## 7) ইন্টারফেস (ছোট) + ডিপেন্ডেন্সি দিক

**কী শিখবে**
- সার্ভিস লেয়ারে `interface` শুধু যা লাগে (Repository মুখ) — টেস্ট/মক সহজ।

**এই প্রজেক্টে**
- `internal/service/item_service.go` — `type ItemRepository interface { ... }`।

**প্র্যাকটিস**
- মেমরিতে বলো: সার্ভিস **কনক্রিট** postgres import করে না, শুধ ইন্টারফেস দেখে।

---

## 8) এরর হ্যান্ডলিং (`errors`, `fmt.Errorf`, Fiber error)

**কী শিখবে**
- `errors.New`, `errors.Is` — sentinel error (`ErrNotFound`)।
- `fmt.Errorf("...: %w", err)` — wrapping, cause chain।
- HTTP লেয়ারে `fiber.NewError(status, msg)` — স্ট্যাটাস কোড explicit।

**এই প্রজেক্টে**
- `internal/service/item_service.go` — `pgx.ErrNoRows` → `ErrNotFound`।
- `internal/transport/http/error_handler.go` — centralized JSON error + `request_id`।

**প্র্যাকটিস**
- নতুন ডোমেইন এরর যোগ করে হ্যান্ডলারে `errors.Is` দিয়ে 404/400 ম্যাপ করো।

---

## 9) JSON + Fiber HTTP হ্যান্ডলার

**কী শিখবে**
- Request body: `c.BodyParser(&struct)`।
- Response: `c.JSON`, `c.Status(...).JSON`।
- Path/query params: `c.Params`, `c.Query`।

**এই প্রজেক্টে**
- `internal/transport/http/handler/auth_handler.go`, `item_handler.go`।

**প্র্যাকটিস**
- একই প্যাটার্নে নতুন `POST` হ্যান্ডলার লিখো।

---

## 10) মিডলওয়্যার চেইন (order matters)

**কী শিখবে**
- `app.Use` order = execution order।
- `c.Next()` — পরের হ্যান্ডলার; return এর পর আবার উপরের মিডলওয়্যারে ফিরে আসে (post-processing)।

**এই প্রজেক্টে**
- `internal/app/app.go`: CORS → RequestIdentity → RequestLogger → routes।
- `router.go`: `Authenticate` গ্রুপ, তারপর per-route `Authorize`।

**প্র্যাকটিস**
- লগিং মিডলওয়্যার কেন `c.Next()` এর আশেপাশে — latency কিভাবে মাপে বুঝো।

---

## 11) JWT অথেন্টিকেশন

**কী শিখবে**
- টোকেন সাইন (`HS256`), claims (role, email), এক্সপায়ারি।
- প্রতি রিকোয়েস্টে `Authorization: Bearer` পার্স।

**এই প্রজেক্টে**
- `internal/service/auth_service.go` — `Login`, `ParseToken`, `Claims`।
- `internal/transport/http/middleware/auth.go` — `Authenticate`।

**প্র্যাকটিস**
- JWT payload decode (jwt.io বা লগ) করে claims দেখো।

---

## 12) RBAC (Authorization)

**কী শিখবে**
- অথেন্টিকেশন = কে; অথরাইজেশন = কী করতে পারে।
- রোল সেট চেক (`admin` vs `user`)।

**এই প্রজেক্টে**
- `middleware.Authorize(service.RoleAdmin, ...)`।

**প্র্যাকটিস**
- নতুন রোল `moderator` যোগ করে একটা রুট শুধু মডারেটর — রাউটার + middleware ধরন বদলাও।

---

## 13) PostgreSQL + `pgxpool`

**কী শিখবে**
- কানেকশন পুল — concurrency তে থ্রেড/গোরুটিন প্রতি কানেকশন খুলে রাখা ভালো নয়।
- `Ping`, pool limits, idle/lifetime — অপারেশন স্টেবিলিটি।

**এই প্রজেক্টে**
- `internal/platform/db/postgres.go` — `pgxpool.ParseConfig`, simple protocol (PgBouncer-friendly), ping retry।

**প্র্যাকটিস**
- `PG_MAX_CONNS` কমিয়ে লোড টেস্ট মেন্টালি চিন্তা করো।

---

## 14) `sqlc` ওয়ার্কফ্লো

**কী শিখবে**
- SQL সোর্স `db/query` — স্কিমা `db/schema` — `sqlc generate` — টাইপড Go।
- হাতে generated ফাইল এডিট না করা।

**এই প্রজেক্টে**
- `sqlc.yaml`, `db/query/items.sql`, `internal/repository/postgres/sqlc/*.go`।
- `item_repository.go` — adapter: sqlc মডেল → `domain.Item`।

**প্র্যাকটিস**
- নতুন `SELECT` কুয়েরি যোগ → `make sqlc` → রিপোজিটরিতে মেথড wrap।

---

## 15) ডোমেইন মডেল ও ভ্যালিডেশন

**কী শিখবে**
- HTTP স্ট্রাক্ট আলাদা, core entity আলাদা — রিইউজ ও টেস্ট সহজ।

**এই প্রজেক্টে**
- `internal/domain/item.go` — `ValidateForCreate`।

**প্র্যাকটিস**
- `title` max length বদলালে শুধ ডোমেইন vs হ্যান্ডলার — কোথায় রাখা উচিত ঠিক করো।

---

## 16) গোরুটিন, চ্যানেল, `select`, `sync.WaitGroup`

**কী শিখবে**
- Background worker pool pattern।
- `select` + `ctx.Done()` — শাটডাউন।
- `close(ch)` + `wg.Wait()` — ড্রেন।
- ননব্লকিং সেন্ড (`default`) — ব্যাকপ্রেশারে ড্রপ।

**এই প্রজেক্টে**
- `internal/platform/async/audit_logger.go` — পুরোটাই পড়ার মতো উদাহরণ।

**প্র্যাকটিস**
- ওয়ার্কার সংখ্যা/বাফার সাইজ বদলে বিহেভিয়ার ভাবো।

---

## 17) স্ট্রাকচার্ড লগ (`log/slog`)

**কী শিখবে**
- key-value লগ, JSON handler — লগ অ্যাগ্রিগেটরে পাঠানো সহজ।

**এই প্রজেক্টে**
- `cmd/api/main.go` — root logger।
- `middleware/observability.go` — per-request log।
- `error_handler.go` — error log।

**প্র্যাকটিস**
- একটা নতুন ইভেন্টে একই স্টাইলে `logger.Info("event", "key", value)` যোগ করো।

---

## 18) কনফিগ (`os.Getenv`, validation)

**কী শিখবে**
- 12-factor style: পোর্ট, DSN, সিক্রেট env থেকে।
- required env missing হলে fail fast (`log.Fatalf` বা error return)।

**এই প্রজেক্টে**
- `internal/config/config.go`।

**প্র্যাকটিস**
- নতুন `FEATURE_X_ENABLED` env যোগ করে হ্যান্ডলারে ব্রাঞ্চ করো।

---

## 19) মাইগ্রেশন vs `sqlc` স্কিমা (দুটোর ভূমিকা)

**কী শিখবে**
- **মাইগ্রেশন** — রিয়েল DB তে টেবিল তৈরি/আপডেট (`migrations/` + `make migrate`)।
- **`db/schema`** — শুধ sqlc-এর টাইপচেক/জেনারেশন; রানটাইমে অটো অ্যাপ্লাই হয় না।

**এই প্রজেক্টে**
- `migrations/001_init.sql` এবং `db/schema/items.sql` — একই টেবিলের ধারণা; চেঞ্জ করলে দুটো সিঙ্ক রাখতে হবে (বা পরে এক সোর্সে নিয়ে যাওয়া যায়)।

---

## 20) ডেভ এক্সপেরিয়েন্স (Make, Air)

**কী শিখবে**
- `Makefile` — এক কমান্ডে বিল্ড/রান/জেনারেট।
- Air — ফাইল চেঞ্জে রিবিল্ড।

**এই প্রজেক্টে**
- `Makefile`, `.air.toml`।

---

## নিজের প্রজেক্ট শুরু করতে “মিনিমাম চেকলিস্ট”

তুমি নিচেরগুলো ছাড়া ছাড়া না করলে বেশিরভাগ API সার্ভিস দাঁড়াবে:

1. `main` + `internal/app` (বা এক জায়গায়) wire-up  
2. `context` + graceful shutdown  
3. `defer` দিয়ে DB close  
4. রাউটার + একটা হ্যান্ডলার  
5. সার্ভিস + রিপোজিটরি ইন্টারফেস  
6. DB pool + মাইগ্রেশন  
7. (optional কিন্তু শক্তিশালী) `sqlc`  
8. সেন্ট্রাল এরর + স্ট্রাকচার্ড লগ  
9. হেলথ চেক (`/healthz`, `/readyz`)  
10. অথ + অথর (JWT + role) যদি ইউজার থাকে  

---

## আরও গভীরে যেতে চাইলে (এই রিপোর বাইরে)

এগুলো এই প্রজেক্টে এখনো পূর্ণ নয়, কিন্তু নিজের প্রজেক্টে লাগবে:

- টেস্ট: `testing`, table-driven tests, httptest, integration + testcontainers  
- মেট্রিক্স/ট্রেসিং: Prometheus, OpenTelemetry  
- রেট লিমিট, আইডempotency, আউটবক্স প্যাটার্ন  

---

## সম্পর্কিত ডক

- পুরো প্রজেক্ট গাইড: [`DOCS.md`](./DOCS.md)  
- রিকোয়েস্ট লাইফসাইকেল: [`REQUEST_LIFECYCLE.md`](./REQUEST_LIFECYCLE.md)  
