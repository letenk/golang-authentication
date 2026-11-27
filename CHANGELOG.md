# Changelog

All notable changes to this boilerplate. Versions are git tags; see "Using This Boilerplate" in the README for how a project created from it pulls in a new version.

Versioning: versions stay on `v1` (Go needs a `/v2` module path for `v2+` tags, which a template gains nothing from). A release that breaks existing projects says so under **Breaking** and lists what to do under **Upgrade notes**.

## [v1.1.0] - 2026-11-27

### Breaking

- **Startup checks.** The app refuses to start when `APPLICATION_ENV` is not exactly `development`, `staging` or `production`; when `APPLICATION_ENV=production` and `APPLICATION_CORS_ALLOWED_ORIGINS` is empty or contains `*`; or when PostgreSQL has `standard_conforming_strings` off.
- **Status codes.** Server-side failures (`DataGetFailed`, `DataCreateFailed`, `DataUpdateFailed`, `DataDeleteFailed`) return 500 instead of 422. Every 500 body now carries only `"Internal Server Error"`; the real error is logged on the server. Client errors (`DataConflict`, `InvalidArgument`, ...) stay 422.
- **Login response** no longer includes user data; it returns tokens only. Use `GET /auth/me` for the user.
- **Phone numbers** are validated as E.164 (`+628123456789`) for every country, replacing the Indonesia-only normalizer.
- **OTPs** live in one generic `otps` table. Migration `20260509000000_create_generic_otps_table` drops `password_reset_otps` and `email_verification_otps`.
- **Persistence moved from Bob to Bun.** Repositories were rewritten, the transaction service passes the transaction through `context` (`WithTx(ctx, fn)`), and the `WithExec` variants are gone. `internal/archtest` forbids persistence imports outside repositories and `configs/database`.
- **`PUT /user/profile`** returns 400 when neither `full_name` nor `phone` is sent.

### Added

- Session management: `GET /user/sessions`, `DELETE /user/sessions/{id}`.
- Profile and password endpoints: `PUT /user/profile`, `PUT /user/password`.
- Forgot and reset password with an emailed OTP (SMTP, sent asynchronously).
- Email verification with OTP: `POST /auth/verify-email`, `POST /auth/resend-verification-email`.
- HttpOnly cookie transport: login and refresh set `access_token` and `refresh_token` cookies; protected routes accept the cookie when no `Authorization` header is sent; refresh and logout read the refresh token from the cookie when the body is empty.
- Swagger UI at `/swagger/index.html` outside production, with request examples; `make swagger` regenerates `api/swagger/`.
- `APPLICATION_CORS_ALLOWED_ORIGINS`: only listed origins may call the API from a browser, with cookies.
- PostgreSQL integration test harness (`make test-db-up`, `make test-integration`).

### Fixed

- 500 responses could expose internal error text such as SQL errors.
- CORS answered `Access-Control-Allow-Origin: *` to every site.
- Two concurrent requests with the same OTP could both succeed; now only one does.
- Two concurrent refresh requests with the same refresh token could both get new tokens; now the second gets 401 (token revoked).
- README: logout needs the refresh token, not an access token.

### Upgrade notes

1. Set `APPLICATION_ENV` to one of the three values in every environment.
2. In production, set `APPLICATION_CORS_ALLOWED_ORIGINS` to the exact frontend origins, comma-separated, without `*`.
3. Make sure PostgreSQL runs with `standard_conforming_strings = on` (the default since PostgreSQL 9.1).
4. Run `make migration-up`. Pending password-reset and email-verification OTPs are dropped with their tables; users just request a new one.
5. Check that stored phone numbers are in E.164 format, or users who log in by phone will not be found.
6. Clients: read user data from `GET /auth/me` after login, and treat 500 (not 422) as "server failed, retry later".
7. Port custom repository code from Bob to Bun, and keep it inside repositories so `internal/archtest` passes.

## [v1.0.5] - 2026-03-08

### Added

- Rate limiting per IP: login 5 requests/minute, refresh 10 requests/minute.

## [v1.0.4] - 2026-03-08

### Added

- Soft delete for users.

## [v1.0.3] - 2026-03-07

### Added

- Unit tests.

## [v1.0.2] - 2026-03-07

### Changed

- Docker uses a single `.env` file.

## [v1.0.1] - 2026-03-07

### Added

- Login by phone number.

## [v1.0.0] - 2026-03-07

First release: register, login with refresh token, refresh, logout, logout from all devices, `GET /auth/me`, structured error responses, Docker setup, Goose migrations, Bob ORM.
