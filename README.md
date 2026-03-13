# go_infra_link

Full-stack project with:

- Go backend in [`backend/`](./backend)
- Svelte web app in [`apps/web-main/`](./apps/web-main)
- Docker Compose for local development

## Quick Start

### Run the full stack with Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

Local endpoints:

- Web app: `http://localhost:5173`
- Backend API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- pgAdmin: `http://localhost:5050`

### Seed data

In another terminal:

```bash
docker compose exec api go run ./cmd/seeder
```

## Environment Setup

The root `.env` is the shared local development config for Compose.

Start from:

```bash
cp .env.example .env
```

Important values:

- `BACKEND_PORT`
- `FRONTEND_PORT`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_DB`
- `JWT_SECRET`
- `SEED_USER_*`

## Current State

The repository already has the main pieces for Docker-based local development:

- `docker-compose.yml` starts `api`, `web-main`, `postgres`, and `pgAdmin`
- backend config loads environment variables from the repo root
- the web app uses the API container through `BACKEND_URL=http://api:8080`
- shared frontend packages now live under `packages/ui-svelte`, `packages/theme`, and `packages/i18n`
- backend feature modules now live under `backend/internal/modules`

There are also some structural issues that should be cleaned up before expanding to a second app:

- this README previously described an outdated backend-only flow
- the current web app still contains app-specific code that can be extracted further
- a shared TypeScript API/client package is still missing
- backend HTTP handlers are still centralized and can be grouped by module later

## Recommended Next Step

Before adding the second website, align the repo around a monorepo structure with shared packages. The proposed direction is documented in [`docs/PROJECT_STRUCTURE.md`](./docs/PROJECT_STRUCTURE.md).

## Swagger

To regenerate Swagger docs:

```bash
cd backend
swag init -g ./cmd/app/main.go -o ./docs
```

## License

See [LICENSE](LICENSE) for details.
