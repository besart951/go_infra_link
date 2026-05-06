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
