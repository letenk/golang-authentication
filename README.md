# Go Auth System — Authentication Boilerplate Built with Production-Grade Patterns

A well-architected authentication boilerplate built with Go, designed to be the starting point for any Go project that needs a solid auth foundation. The patterns and architecture here — type-safe SQL via sqlc, compile-time DI via Wire, token rotation, generic OTP system, internationally-aware phone normalization — are the same patterns you'd reach for in production code.

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
| **Database access** | [sqlc v1.31](https://github.com/sqlc-dev/sqlc) | Type-safe Go generated from plain SQL; generated code has no runtime dependency |
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

## Why sqlc?

Database access is written as **plain SQL** in `db/queries/*.sql`. [sqlc](https://docs.sqlc.dev) reads those queries plus the goose migrations and generates type-safe Go in `db/sqlcgen/`.

- **Compile-time safety** — a renamed column or a wrong parameter type fails `make sqlc-gen` / `go build`, not production.
- **No runtime dependency** — the generated code is plain `database/sql`. If sqlc were abandoned tomorrow, the code keeps working and can even be edited by hand.
- **No database needed to generate** — the schema comes from `migrations/`, so `make sqlc-check` runs in CI without Postgres.
- **Stable** — sqlc is v1 with a stable API and a large user base.

The trade-off is that highly dynamic queries (many optional filters/sorts) are less pleasant to express; use `sqlc.narg()` + `COALESCE` as in `UpdateUserProfile`, or a small query builder for that one query.

### This project moved from Bob to sqlc

The boilerplate originally used [Bob](https://github.com/stephenafamo/bob). It was replaced because Bob is still pre-1.0 (frequent breaking minor releases), largely maintained by one person, and its generated code depends on the Bob runtime. The migration first decoupled services from Bob (repositories now expose only `internal/domain` types), then swapped each repository to sqlc; that swap touched repository files only, and the repository integration tests passed unchanged. See `docs/superpowers/specs/2026-11-15-bob-to-sqlc-design.md`.

---

## Project Structure

```
.
├── db/
│   ├── queries/                 # Hand-written SQL, one file per table (edit these)
│   └── sqlcgen/                 # sqlc generated code (do not edit manually)
├── cmd/
│   └── main.go                  # Entry point
├── configs/
│   ├── credential/              # Config structs + Viper initialization
│   ├── database/                # DB connection, tx-in-context, NULL helpers
│   ├── jwt_config/              # JWT sign/verify logic
│   └── validator/               # Custom Echo validator
├── docs/                        # Documentation and migration guides
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
├── sqlc.yaml                    # sqlc code generation config
├── docker-compose.test.yml      # Disposable Postgres for integration tests
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

---

## API Endpoints

Base path: `/api/v1`

### Auth — `/auth`

| Method | Endpoint | Auth Required | Description |
|--------|----------|:---:|---|
| `POST` | `/auth/register` | — | Register new user |
| `POST` | `/auth/login` | — | Login (email or phone) · rate limited |
| `POST` | `/auth/refresh` | — | Refresh access token · rate limited |
| `POST` | `/auth/logout` | ✓ | Logout current session |
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

### 3. Generate database code

> Run this after changing a migration or a file in `db/queries/`. No database needed.

```bash
make sqlc-gen
```

### 4. Run the server

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

# Fail if generated sqlc code is out of date (CI)
make sqlc-check

# Run tests with coverage report
make test-coverage

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
APPLICATION_ENV=development
APPLICATION_PORT=8080

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
