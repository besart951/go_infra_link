# go_infra_link

Full-stack infrastructure link project with Go backend and SvelteKit frontend.

## Quick Start

### Backend

```bash
cd backend
go run ./cmd/db-bootstrap
go run ./cmd/app
```

`db-bootstrap` applies the project's forward-only schema migrations and records them in `schema_migrations`.

## Development

### Prerequisites

- Go 1.25.6 or higher
- Node.js 24.x with pnpm 10.29.1 for frontend development
- Docker (optional, for containerized deployment)

### Running with Docker

```bash
docker compose up --build
```

The local frontend service uses Vite for developer ergonomics, but the production deployment model is a static SPA served by Caddy.

For Linux or Bash-based local development, use the helper script:

```bash
./scripts/dev.sh start
```

Run `./scripts/dev.sh help` to see the available actions.

### Running Tests

```bash
cd backend
go test ./...
```

### Quality Gates

Run the same checks locally before pushing:

```powershell
.\scripts\ci.ps1
.\scripts\ci.ps1 -Target backend
.\scripts\ci.ps1 -Target frontend
```

```bash
bash scripts/ci.sh
bash scripts/ci.sh backend
bash scripts/ci.sh frontend
```

Backend gates run `go test ./...`, targeted `go test -race` for realtime/concurrency-sensitive packages, `go vet ./...`, `staticcheck`, and `govulncheck`. Frontend gates run `pnpm install --frozen-lockfile`, `pnpm check`, `pnpm test`, `pnpm build`, and the Prettier-based `pnpm lint`.

CI is defined for GitHub Actions and Forgejo Actions with separate backend and frontend jobs plus dependency caching. Docker image builds run only on `main` or manual dispatch to keep pull-request feedback focused. Frontend CI requires Node.js 24.x and pnpm 10.29.1.

## Swagger

### Generate Swagger docs

From the backend folder:

```bash
cd backend
swag init -g ./cmd/app/main.go -o ./docs
```

### View Swagger UI

If `SWAGGER_ENABLED=true`, start the backend and open:

```
http://localhost:8080/swagger/index.html
```

## Architecture

This project follows **Clean/Hexagonal Architecture** principles:

- **Backend**: Go service with layered architecture (domain, application, infrastructure)
- **Frontend**: Static SvelteKit SPA served by Caddy

### Production deployment contract

- The frontend is built with `@sveltejs/adapter-static`
- The frontend container serves static assets and can proxy same-origin `/api/*` to the backend for standalone Compose deployments
- The `server-setup` edge reverse proxy owns live blue/green routing for `/api/*`
- Backend endpoints remain the only source of truth for auth, cookies, CSRF, and authorization
- PostgreSQL runs on the official `postgres:18.3-alpine3.23` image; major upgrades must use dump/restore, not an in-place data directory reuse
- Production config is fail-closed: `JWT_SECRET` must be strong and non-default, `COOKIE_SECURE=true`, `TRUSTED_PROXIES` must name the reverse proxy IP/CIDR, and CORS origins must never use `*`
- Cookie auth is designed for the same-origin SPA: auth cookies are `HttpOnly`, CSRF is a readable double-submit cookie, and `COOKIE_SAME_SITE=strict` is the default
- `COOKIE_DOMAIN` is optional. Leave it empty for host-only cookies; set it only when cookies must be shared across subdomains
- Production database SSL must use `sslmode=require`, `verify-ca`, or `verify-full`. For an intentionally private database link, set `POSTGRES_SSLMODE=disable` together with `DB_ALLOW_UNSAFE_SSLMODE=true`
- HSTS must be emitted by the HTTPS-terminating proxy. The bundled frontend Caddy container listens on `:80`, so it documents HSTS but does not send it there

Do not rely on SvelteKit server hooks or `+server.ts` routes in production unless the frontend is intentionally migrated to a server adapter such as `adapter-node`.

## Utility Commands

Get a CSRF token from a local login:

```powershell
Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/v1/auth/login" -ContentType "application/json" -Body '{"email":"besart_morina@hotmail.com","password":"password"}' -SessionVariable s | Select-Object -ExpandProperty csrf_token
```

Seed snapshot data into the database:

```bash
docker compose exec backend go run ./cmd/seeder
```

## License

See [LICENSE](LICENSE) for details.
