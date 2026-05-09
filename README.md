# Go Auth System — Authentication Boilerplate Built with Production-Grade Patterns

A well-architected authentication boilerplate built with Go, designed to be the starting point for any Go project that needs a solid auth foundation. The patterns and architecture here — type-safe queries via Bob ORM, compile-time DI via Wire, token rotation, generic OTP system, internationally-aware phone normalization — are the same patterns you'd reach for in production code.

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
| **ORM / Query Builder** | [Bob v0.43](https://github.com/stephenafamo/bob) | Type-safe SQL, code generation, no magic |
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

## Why Bob ORM?

Most Go projects reach for GORM or sqlx. Here's why Bob was chosen instead.

### The Problem with GORM

GORM is convenient but trades correctness for magic:
- Queries are built at runtime via reflection — typos and wrong field names only fail at runtime
- Hook system and association loading have surprising behavior under transactions
- Hard to write custom or complex queries without dropping into raw SQL

### The Problem with sqlx

sqlx is closer to the metal but still requires writing SQL strings by hand — no compile-time guarantees, and no code generation.

### Why Bob

Bob takes a different approach: **code generation from the actual database schema**.

Run `make bob-gen` after every migration, and Bob generates typed model structs, type-safe query builders, and test factories directly from your live schema. If a column is renamed in a migration and the generated code is not updated, the code **won't compile** — the bug is caught before it ever reaches runtime.

**Summary:**

| | GORM | sqlx | Bob |
|---|---|---|---|
| Type safety | Runtime | None | Compile-time |
| Code generation | No | No | Yes |
| Custom queries | Awkward | Raw SQL | Builder API |
| Transaction support | Hooks | Manual | Native executor |
| Test factories | Manual | Manual | Generated |

Bob requires an upfront `make bob-gen` step after every migration, but the payoff is fewer runtime surprises and better IDE support.

For full documentation and usage examples, see the [Bob official documentation](https://bob.stephenafamo.com).

---

## Project Structure

```
.
├── bob/                          # Bob ORM generated code (do not edit manually)
│   ├── models/                  # Generated model structs and query builders
│   ├── factory/                 # Generated test data factories
│   ├── dberrors/                # Per-table database error handling
│   └── dbinfo/                  # Schema metadata
├── cmd/
│   └── main.go                  # Entry point
├── configs/
│   ├── credential/              # Config structs + Viper initialization
│   ├── database/                # DB connection setup
│   ├── jwt_config/              # JWT sign/verify logic
│   └── validator/               # Custom Echo validator
├── docs/                        # Documentation and migration guides
├── exceptions/                  # Domain-level error types
├── internal/
│   ├── adapter/rest/            # Echo server setup, middleware, routing
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
├── bobgen.yaml                  # Bob code generation config
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

### 3. Generate Bob ORM models

> Run this after every migration.

```bash
make bob-gen
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
# Run tests
make test

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

**Repository pattern** — all database access is behind interfaces. This makes unit testing straightforward (swap the real repo for a mock) and keeps the service layer database-agnostic.

**TrxService** — a `WithTx(ctx, fn)` abstraction wraps database transactions. Service methods that need atomicity pass a `bob.Executor` into the callback, keeping transaction management out of the repository layer.

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
