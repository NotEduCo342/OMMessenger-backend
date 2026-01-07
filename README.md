# OMMessenger-backend

OM Messenger backend (Go + Fiber + Postgres).
70% done
This repo currently implements cookie-based auth suitable for intranet deployments:
- Short-lived access JWT in `om_access` (HttpOnly)
- Long-lived refresh token in `om_refresh` (HttpOnly, stored hashed in DB)
- CSRF protection using `om_csrf` + `X-OM-CSRF` header

## Key Endpoints

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/csrf` (issues/rotates CSRF cookie)
- `POST /api/auth/refresh` (requires CSRF)
- `POST /api/auth/logout` (requires CSRF)
- `GET /api/users/me`

## Environment Variables

**Required**
- `JWT_SECRET`: signing secret for access JWTs (server will refuse to start if missing)
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`, `DB_SSLMODE`

**Browser security**
- `ALLOWED_ORIGINS`: comma-separated allow-list for browser `Origin` checks and CORS
- `COOKIE_SECURE`: set to `true` behind TLS (`https://` / `wss://`)
- `COOKIE_DOMAIN`: optional cookie domain
- `COOKIE_SAMESITE`: `Lax` (default) | `Strict` | `None`

**CSRF**
- `CSRF_MODE`: `token` (default) | `origin` | `off`
	- `token`: requires `X-OM-CSRF` to match `om_csrf` cookie for unsafe methods
	- `origin`: only enforces `ALLOWED_ORIGINS`

**Limits**
- `PASSWORD_MIN_LENGTH` (default: `10`)
- `MAX_MESSAGE_LENGTH` (default: `4000`)

## How to build
### Windows:
```shell
go build -o bin/server.exe -v .\cmd\server
```
### Linux:
```shell
go build -o bin/server -v ./cmd/server
```
