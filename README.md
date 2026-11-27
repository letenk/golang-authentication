# Go Auth System — Authentication Boilerplate Built with Production-Grade Patterns

A well-architected authentication boilerplate built with Go, designed to be the starting point for any Go project that needs a solid auth foundation. The patterns and architecture here — an SQL-first query builder kept behind repository interfaces, compile-time DI via Wire, token rotation, generic OTP system, internationally-aware phone normalization — are the same patterns you'd reach for in production code.

This is a **foundation, not a finished production system**. Use it to skip the boilerplate; harden it for your production environment as you grow.

> Built as a personal boilerplate and portfolio project by [Rizky Darmawan](https://github.com/letenk).

---

## Why This Exists

Every new project needs authentication. Instead of rebuilding the same patterns from scratch, this boilerplate provides a well-structured, tested, and extensible auth system that can be dropped into any Go project.

**What's included out of the box:**
- Register, login (email or phone), logout, refresh token
- Multi-device session management
- Email verification via OTP
- Forgot & reset password via OTP
- User profile management
- Rate limiting on sensitive endpoints
- Soft delete support
- JWT authentication with configurable expiry
- Structured error handling
- Dockerized with health checks

---

## Tech Stack

| Layer | Library | Why |
|---|---|---|
| **HTTP Framework** | [Echo v4](https://github.com/labstack/echo) | Minimal, fast, great middleware ecosystem |
| **Database access** | [Bun v1.3](https://github.com/uptrace/bun) | SQL-first query builder on `database/sql`; handles dynamic filters/sorts; dialects for PostgreSQL, MySQL, SQLite, MSSQL |
| **Database** | PostgreSQL | Battle-tested, feature-rich relational DB |
| **Migrations** | [Goose v3](https://github.com/pressly/goose) | Simple, SQL-first migration tool |
| **JWT** | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) | Standard JWT library with full RFC support |
| **Config** | [Viper](https://github.com/spf13/viper) | Flexible config from env, file, or both |
| **Dependency Injection** | [Wire](https://github.com/google/wire) | Compile-time DI, no runtime reflection |
| **Validation** | [go-playground/validator v10](https://github.com/go-playground/validator) | Declarative struct validation via tags |
| **Email** | [gomail.v2](https://gopkg.in/gomail.v2) | Lightweight SMTP email sender |
| **Testing** | [Testify](https://github.com/stretchr/testify) + [mockery](https://github.com/vektra/mockery) | Assertions + auto-generated mocks |
| **Fake Data** | [jaswdr/faker v2](https://github.com/jaswdr/faker) | Realistic test data generation |

---

## Why Bun?

Repositories build queries with [Bun](https://bun.uptrace.dev), an SQL-first query builder that runs on top of the standard `database/sql` pool:

```go
err := r.db.NewSelect(ctx).Model(row).
    Where("email = ?", email).
    Where("deleted_at IS NULL").
    Scan(ctx)
```

- **Dynamic queries are easy** — optional filters and client-chosen sorting are plain `if` + `.Where(...)` / `.Order(...)`, which static-SQL generators handle poorly.
- **Stable** — Bun is v1 with a stable API.
- **Built on `database/sql`** — transactions stay `*sql.Tx` (see `TrxService`), so the transaction layer does not depend on Bun.
- **Multi-dialect** — only PostgreSQL is used today, but placeholders are always `?` and timestamps use `CURRENT_TIMESTAMP`, so adding another database means a new dialect, migrations and an error translator, and rewriting only the few queries that use PostgreSQL-specific SQL. Today that is `RETURNING` in `Create` and `UpdateProfile` (MySQL has no `RETURNING`; those become INSERT/UPDATE followed by SELECT).

The trade-off: column names are checked when a query runs, not at build time. The repository integration tests (`make test-integration`) are therefore mandatory — every repository method is covered by one.

### Keeping Bun contained

- Bun is imported only by `configs/database` and the repository packages. `internal/archtest` (part of `make test`) fails if a service, controller, DTO or `internal/domain` imports Bun or `database/sql`, and also rejects `bun`/`db`/`gorm` struct tags on `internal/domain` types.
- Each repository has its own unexported row struct (`userRow`, `otpRow`, ...) and maps it to the domain type. Services only ever see `internal/domain`.
- Build queries only through `db.NewSelect/NewInsert/NewUpdate/NewDelete/NewRaw(ctx)`; they join the transaction carried by `ctx`. The `*bun.DB` is unexported, so a repository cannot build a Bun query that skips the transaction. `db.DB` (the raw pool, used by `TrxService` and the health check) is still exported, so repositories must not call it directly — it would run outside the caller's transaction.
- Bun substitutes `?` values into the SQL text on the client and escapes strings by doubling single quotes. That is safe only while PostgreSQL's `standard_conforming_strings` is `on` (the default), so `NewPostgres` refuses to start when it is off.
- The schema belongs to the goose migrations in `migrations/`; Bun never creates or alters tables.

### Filtering and sorting (pattern for list endpoints)

No endpoint lists with filters yet. When the first one is added, follow this shape so the criteria stay ORM-agnostic and safe:

```go
// internal/domain
type UserSortField string

const (
    UserSortCreatedAt UserSortField = "created_at"
    UserSortName      UserSortField = "name"
)

type UserFilter struct {
    Name   *string
    SortBy UserSortField
    Desc   bool
    Limit  int
    Offset int
}
```

- The service passes a `domain.UserFilter`; Bun query builders never leave the repository.
- The repository maps `UserSortField` to a column through a whitelist and rejects anything else with an error. Never pass request input to `.Order()`: Bun quotes the identifier but accepts **any** column (sorting by `password` would leak information about hashes) and drops an invalid direction with only a log line, not an error.
- Filters always use placeholders, never string concatenation. For a substring search, escape LIKE wildcards in the input first, otherwise a user can send `%` or `_` to match everything:

  ```go
  pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(*f.Name) + "%"
  q = q.Where(`LOWER(name) LIKE LOWER(?) ESCAPE '\'`, pattern)
  ```

  Prefer portable SQL over PostgreSQL-only operators such as `ILIKE`. Cap `Limit` in the repository.

### Replacing the ORM or the database

1. Make sure `make test-integration` is green on the current implementation.
2. Rewrite one repository at a time; keep its interface and its `*_integration_test.go` file unchanged.
3. Re-run the same integration tests — passing unchanged is the proof the behaviour is the same.
4. Run `make test` so `internal/archtest` confirms nothing leaked out of the repositories.
5. Remove the old library (`go mod tidy`).

Changing the database additionally needs a Bun dialect in `configs/database`, migrations for that database, and a unique-violation translator in `configs/database/errors.go`.

### History: Bob → sqlc → Bun

The boilerplate originally used [Bob](https://github.com/stephenafamo/bob). It was replaced because Bob is still pre-1.0 (frequent breaking minor releases), largely maintained by one person, and its generated code depends on the Bob runtime. The migration first decoupled services from Bob (repositories now expose only `internal/domain` types), then swapped each repository to sqlc; that swap touched repository files only, and the repository integration tests passed unchanged. sqlc was later replaced by Bun because upcoming list endpoints need dynamic filtering and sorting, which sqlc's static queries express poorly; again only repositories and `configs/database` changed, and the integration tests passed unchanged.

---

## Project Structure

```
.
├── api/
│   └── swagger/                 # Generated Swagger spec (make swagger); do not edit by hand
├── cmd/
│   └── main.go                  # Entry point
├── configs/
│   ├── credential/              # Config structs + Viper initialization
│   ├── database/                # DB connection, Bun query constructors, tx-in-context
│   ├── jwt_config/              # JWT sign/verify logic
│   └── validator/               # Custom Echo validator
├── exceptions/                  # Domain-level error types
├── internal/
│   ├── adapter/rest/            # Echo server setup, middleware, routing
│   ├── domain/                  # Entities + domain errors (stdlib only)
│   ├── testutil/pgtest/         # Real-Postgres helper for integration tests
│   └── applications/
│       ├── auth/                # Auth feature (controller, service, dto, mocks)
│       ├── email/               # SMTP email service
│       ├── otp/                 # Generic OTP repository (purpose: email_verification, password_reset, ...)
│       ├── refresh_token/       # Refresh token repository
│       ├── transaction/         # TrxService — WithTx abstraction
│       └── user/                # User repository + profile controller
├── middleware/
│   └── authentication/          # JWT middleware (Authenticate)
├── migrations/                  # SQL migration files (Goose)
├── .env-example                 # Environment variable template
├── docker-compose.test.yml      # Disposable Postgres for integration tests
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

---

## API Endpoints

Base path: `/api/v1`

### Interactive docs (Swagger)

With `APPLICATION_ENV` set to `development` or `staging`, open `http://localhost:8080/swagger/index.html` (use your `APPLICATION_PORT`). In `production` the route is not registered and returns `404`.

For protected endpoints, call `POST /auth/login`, click **Authorize** and enter `Bearer <access_token>`. Login also sets the `access_token` cookie in the browser, and the API accepts it when no `Authorization` header is sent, so protected calls from Swagger UI work even before you click **Authorize**. Call `POST /auth/logout` or clear cookies to test the header path alone.

The spec is generated from comments on the handlers. After changing an endpoint or its annotations, run `make swagger` and commit `api/swagger/`.

### Auth — `/auth`

| Method | Endpoint | Auth Required | Description |
|--------|----------|:---:|---|
| `POST` | `/auth/register` | — | Register new user |
| `POST` | `/auth/login` | — | Login (email or phone) · rate limited |
| `POST` | `/auth/refresh` | — | Refresh access token · rate limited |
| `POST` | `/auth/logout` | — | Logout current session · needs the refresh token (body or `refresh_token` cookie), not an access token |
| `POST` | `/auth/logout-all` | ✓ | Logout all devices |
| `GET` | `/auth/me` | ✓ | Get current user |
| `DELETE` | `/auth/me` | ✓ | Delete account |
| `POST` | `/auth/forgot-password` | — | Send OTP to email |
| `POST` | `/auth/reset-password` | — | Reset password with OTP |
| `POST` | `/auth/verify-email` | ✓ | Verify email with OTP |
| `POST` | `/auth/resend-verification-email` | ✓ | Resend verification OTP |

### User — `/user`

| Method | Endpoint | Auth Required | Description |
|--------|----------|:---:|---|
| `GET` | `/user/sessions` | ✓ | List all active sessions |
| `DELETE` | `/user/sessions/:id` | ✓ | Revoke a specific session |
| `PUT` | `/user/profile` | ✓ | Update name / phone |
| `PUT` | `/user/password` | ✓ | Change password |

### Health

| Method | Endpoint | Description |
|--------|----------|---|
| `GET` | `/health` | Health check with DB connectivity |

---

## Using This Boilerplate

Releases are git tags (`v1.1.0`, `v1.2.0`, ...). [CHANGELOG.md](CHANGELOG.md) lists what changed in each one, with upgrade notes for breaking changes. Versions stay on `v1`: Go only accepts a `v2+` tag when the module path ends in `/v2`, and a template that nobody imports gains nothing from that rename.

A boilerplate is copied, not imported, so a new project does not follow new versions automatically. The steps below keep an untouched copy of the boilerplate on a `boilerplate` branch, so every update is an ordinary `git merge`: files you never changed update cleanly, and files you customized show normal conflict markers.

The examples use `github.com/you/your-app` as the new module path and `v1.1.0` → `v1.2.0` as versions.

### Start a new project

```bash
go run golang.org/x/tools/cmd/gonew@latest \
  github.com/letenk/golang-authentication@v1.1.0 github.com/you/your-app
cd your-app

# gonew rewrites Go import paths only; two files hold the old path as plain text
sed -i.bak 's#github.com/letenk/golang-authentication#github.com/you/your-app#g' \
  .mockery.yaml internal/archtest/layering_test.go && rm -f .mockery.yaml.bak internal/archtest/layering_test.go.bak

git init -b main
git add -A && git commit -m "chore: start from boilerplate v1.1.0"
git branch boilerplate          # untouched copy, used for future updates
```

Skipping the `sed` step does not break the build, but `internal/archtest` then silently stops checking imports of `configs/database`.

Then make it your project: `cp .env-example .env` and fill it in (`APPLICATION_CORS_ALLOWED_ORIGINS` is required in production), change `APPLICATION_NAME`, the Swagger `@title` in `cmd/main.go` (then `make swagger`), and this README. Run `make test`.

### Update to a newer boilerplate version

Read the version's entry in `CHANGELOG.md` first, especially its upgrade notes. Then:

```bash
# 1. Put the new version on the boilerplate branch
git switch boilerplate
find . -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +
go run golang.org/x/tools/cmd/gonew@latest \
  github.com/letenk/golang-authentication@v1.2.0 github.com/you/your-app ../boilerplate-tmp
cp -R ../boilerplate-tmp/. . && rm -rf ../boilerplate-tmp
sed -i.bak 's#github.com/letenk/golang-authentication#github.com/you/your-app#g' \
  .mockery.yaml internal/archtest/layering_test.go && rm -f .mockery.yaml.bak internal/archtest/layering_test.go.bak
git add -A && git commit -m "chore: boilerplate v1.2.0"

# 2. Merge it into your project on a branch, resolve conflicts, test, open a PR
git switch main
git switch -c chore/boilerplate-v1.2.0
git merge boilerplate
make test
```

Never commit your own changes to the `boilerplate` branch; it must stay identical to the boilerplate release so that git can tell your changes from upstream ones.

### Database migrations from the boilerplate

Migration files are named by timestamp, so a migration that arrives in an update can be older than migrations your project already applied. `make migration-up` handles this: the bundled migration binary runs Goose with `WithAllowMissing`, which applies such older files. If you run the `goose` CLI directly instead, pass `-allow-missing`. Check the upgrade notes for migrations that drop or rewrite data.

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 16
- Docker & Docker Compose (optional)

### 1. Clone and configure

```bash
git clone https://github.com/letenk/golang-auth-system.git
cd golang-auth-system
cp .env-example .env
# Edit .env with your database and SMTP credentials
```

### 2. Run database migrations

```bash
make migration-up
```

### 3. Run the server

```bash
make run
```

### Run with Docker

```bash
docker compose up -d
```

---

## Development

```bash
# Run unit tests
make test

# Run unit + integration tests against a real Postgres
make test-db-up          # disposable Postgres on :55432 (or set TEST_DATABASE_URL)
make test-integration
make test-db-down

# Run tests with coverage report
make test-coverage

# Regenerate Swagger docs after changing handler annotations
make swagger

# Create a new migration
make migration-create name=add_roles_table type=sql

# Rollback last migration
make migration-down
```

---

## Environment Variables

Copy `.env-example` to `.env` and fill in the values.

```env
# Application
APPLICATION_NAME=auth-system
APPLICATION_ENV=development             # development | staging | production (anything else stops the app)
APPLICATION_PORT=8080
APPLICATION_CORS_ALLOWED_ORIGINS=http://localhost:3000   # comma-separated; required in production, no "*"

# Database
DB_CONFIGS_HOST=localhost
DB_CONFIGS_PORT=5432
DB_CONFIGS_USER=postgres
DB_CONFIGS_PASSWORD=your_password
DB_CONFIGS_NAME=authentication_db
DB_CONFIGS_SSLMODE=disable

# JWT
AUTH_JWT_SECRET=your-super-secret-jwt-key
AUTH_JWT_ACCESS_TOKEN_EXPIRE=15m
AUTH_JWT_REFRESH_TOKEN_EXPIRE=7d

# OTP
AUTH_OTP_EXPIRE=5m
AUTH_OTP_LENGTH=5

# Email SMTP
# Dev: Mailtrap (https://mailtrap.io) or Mailhog
# Prod: Gmail SMTP, SendGrid, Mailgun, etc.
EMAIL_SMTP_HOST=sandbox.smtp.mailtrap.io
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USERNAME=your_username
EMAIL_SMTP_PASSWORD=your_password
EMAIL_SMTP_FROM=noreply@your-domain.com
```

---

## Database Schema

```
users
  id, name, email, phone, password
  is_verified, verified_at
  login_type, deleted_at
  created_at, updated_at (with audit trail)

refresh_tokens
  id, user_id, token
  device_name, device_id, ip_address, user_agent
  expires_at, revoked_at, replaced_by_token

otps
  id, user_id, code, purpose, expires_at, used_at
  -- single generic table, `purpose` distinguishes use case
  -- ('email_verification', 'password_reset', ...)
```

---

## Architecture Decisions

**Clean Architecture** — each feature lives in `internal/applications/<feature>/` with its own controller, service, repository, and DTO layers. Dependencies flow inward: controller → service → repository.

**Repository pattern** — all database access is behind interfaces that accept and return only `internal/domain` types (or the repository's own params structs). ORM/generated types never leave the repository package, so the service layer is database-agnostic and the persistence library can be swapped by rewriting repositories alone.

**TrxService** — `WithTx(ctx, func(ctx context.Context) error)` starts a transaction and passes a context that carries it. Repository calls made with that context join the transaction automatically; nested `WithTx` calls join the outer one.

**Domain errors** — repositories return `domain.ErrNotFound`, and translate PostgreSQL unique violations into `domain.ErrEmailTaken` / `domain.ErrPhoneTaken`, so a race between two registrations (or an email still held by a soft-deleted account) answers 422 `DataConflict` (10012) instead of 500.

**Async email** — all email sends (OTP delivery) run in a background goroutine with an independent context and timeout. This keeps HTTP response times fast regardless of SMTP latency.

**Rate limiting** — login and refresh endpoints are rate limited per IP to slow down brute-force attacks.

**Soft delete** — users are soft-deleted (`deleted_at`). All queries filter `WHERE deleted_at IS NULL` so deleted accounts cannot log in.

---

## Roadmap

| Feature | Status |
|---------|--------|
| Register, Login (email + phone) | ✅ Done |
| Refresh token with rotation | ✅ Done |
| Logout / Logout all devices | ✅ Done |
| Multi-device session management | ✅ Done |
| User profile & password update | ✅ Done |
| Email OTP verification | ✅ Done |
| Forgot & reset password | ✅ Done |
| Rate limiting (login, refresh) | ✅ Done |
| Soft delete filter | ✅ Done |
| Service & handler unit tests | ✅ Done |
| Role system (RBAC) | 🔲 Planned |
| JWT RS256 (asymmetric) | 🔲 Planned |
| Cleanup expired tokens (cron) | 🔲 Planned |

---

## License

MIT License

Copyright (c) 2026 Rizky Darmawan

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
