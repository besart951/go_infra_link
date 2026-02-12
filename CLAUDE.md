# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Full-stack infrastructure management application. Go backend (Gin + GORM) with SvelteKit frontend (Svelte 5 + Tailwind CSS + bits-ui).

## Commands

### Backend (`cd backend`)

| Task           | Command                                       |
| -------------- | --------------------------------------------- |
| Run server     | `go run ./cmd/app`                            |
| Build          | `go build ./cmd/app`                          |
| Run tests      | `go test ./...`                               |
| Run with coverage | `go test -cover ./...`                     |
| Lint           | `make lint`                                   |
| Lint + fix     | `make lint-fix`                               |
| Migrate up     | `make migrate-up`                             |
| Migrate down   | `make migrate-down`                           |
| Swagger docs   | `swag init -g ./cmd/app/main.go -o ./docs`    |
| Seed data      | `go run ./cmd/seeder`                         |

### Frontend (`cd frontend`)

| Task               | Command                |
| ------------------ | ---------------------- |
| Dev server         | `pnpm run dev`         |
| Build              | `pnpm run build`       |
| Type check         | `pnpm run check`       |
| Type check (watch) | `pnpm run check:watch` |
| Format             | `pnpm run format`      |
| Lint (Prettier)    | `pnpm run lint`        |

### Full Stack

`docker-compose up` from root starts all services:
- **backend** (port 8080) - Go API server with hot reload
- **frontend** (port 5173) - SvelteKit dev server
- **postgres** (port 5432) - PostgreSQL 16 database

Useful commands:
- `docker compose exec backend go run ./cmd/seeder` - Seed database
- `docker compose logs -f backend` - View backend logs
- `docker compose down -v` - Stop and remove volumes

## Architecture

### Backend (`backend/`)

**Hexagonal/Clean Architecture** with manual dependency injection.

```
cmd/app/main.go          → Entry point
internal/
  domain/                → Entities, repository interfaces, domain errors (no external deps)
    auth/, user/, team/, project/, facility/
  repository/            → GORM implementations of domain repository interfaces
  service/               → Business logic (auth, user, team, project, facility, rbac, admin)
  handler/               → Gin HTTP handlers + DTOs + mappers
    middleware/           → Auth (JWT), CSRF, RBAC, AccountStatus guards
    dto/                 → Request/response data transfer objects
    mapper/              → Entity ↔ DTO conversion
  wire/                  → Manual DI wiring (repositories.go → services.go → handlers.go)
  config/                → Environment-based configuration
  db/                    → Database connection (supports PostgreSQL, MySQL)
```

**Key patterns:**

- REST API at `/api/v1/*`, Swagger at `/swagger/index.html`
- JWT access tokens (15min default) + refresh tokens in cookies
- CSRF double-submit cookie pattern for state-changing requests
- **Migrations**: Supports both AutoMigrate (GORM auto-creates/updates schema on startup, default for dev) and manual SQL migrations (`backend/migrations/` via `make migrate-up` for production)
- Adding a feature: domain entity → repository interface → repository impl → service → handler → routes → wire

**Role hierarchy (7-tier):** `superadmin(100)` > `admin_fzag(90)` > `fzag(80)` > `admin_planer(70)` > `planer(60)` > `admin_entrepreneur(50)` > `entrepreneur(40)`

### Frontend (`frontend/`)

**SvelteKit 2 + Svelte 5** with file-based routing, runes-based reactivity.

```
src/
  lib/
    api/                 → HTTP client layer (client.ts handles CSRF + errors)
    domain/entities/     → TypeScript types mirroring backend domain
    stores/              → Svelte 5 rune stores (.svelte.ts files)
      auth.svelte.ts     → Auth state, role checks, allowed roles
      list/              → Generic paginated list store pattern
    components/
      ui/                → shadcn-svelte components (bits-ui based)
      facility/          → Facility-specific form components
      list/              → PaginatedList reusable table component
    utils/
      permissions.ts     → ROLE_LABELS, ROLE_PERMISSIONS, canPerform()
    hooks/               → Custom Svelte hooks (useFormState, etc.)
  routes/
    (app)/               → Protected routes (require auth)
      users/, teams/, facility/, projects/, account/
    (auth)/              → Public routes (login)
```

**Key patterns:**

- **API Proxy**: Frontend proxies all `/api/v1/*` requests to backend via SvelteKit catch-all route (`routes/api/v1/[...path]/+server.ts`). Backend URL configurable via `BACKEND_URL` env var (defaults to `localhost:8080`)
- **SPA Mode**: Uses `@sveltejs/adapter-static` with fallback routing for client-side navigation
- shadcn-svelte component library via bits-ui (headless) + Tailwind CSS
- DropdownMenu trigger uses `{#snippet child({ props })}` pattern (see `nav-user.svelte`)
- Popover+Command combo for searchable selects (see `AsyncCombobox.svelte`)
- `PaginatedList.svelte` + entity stores for all list pages
- API client auto-extracts CSRF token from cookies, sends with mutations
- State via Svelte 5 `$state`, `$derived`, `$effect` runes (NOT legacy `$:` or writable stores)
- Domain types in `src/lib/domain/entities/` mirror backend types

### Frontend Component Conventions

- shadcn components imported as namespace: `import * as Dialog from '$lib/components/ui/dialog/index.js'`
- Badge variants: `default`, `secondary`, `destructive`, `outline`, `success`, `warning`
- Use `.js` extension in import paths for shadcn components; non-`.js` imports for `$lib/api/*` and `$lib/utils/*` are a pre-existing inconsistency (both work at runtime)
- Toast: `addToast(message: string, type: 'success' | 'error')`
- Confirm dialog: `const ok = await confirm({ title, message, confirmText, cancelText, variant })`

## Environment

**Backend** (`.env` in root, see `backend/example.env`):
- `DB_DRIVER`: Database type (`postgres` or `mysql`)
- `DB_DSN`: Database connection string
- `JWT_SECRET`: Secret for signing JWT tokens
- `ACCESS_TOKEN_TTL`: Access token lifetime (default: `15m`)
- `REFRESH_TOKEN_TTL`: Refresh token lifetime (default: `720h` / 30 days)
- `HTTP_ADDR`: Server address (default: `:8080`)
- `COOKIE_SECURE`: Set `true` in production for HTTPS-only cookies
- Default dev DB: PostgreSQL

**Frontend**:
- `BACKEND_URL`: Backend API URL for server-side proxy
  - Development: defaults to `http://localhost:8080`
  - Docker: set to `http://backend:8080` in docker-compose.yml

## Pre-existing Issues

The following build/check issues pre-date current work and are not regressions:

- `svelte-check` reports module resolution errors for non-`.js` imports (e.g., `$lib/api/users` instead of `$lib/api/users.js`)
- `FieldDeviceForm.svelte` has a Svelte 4 reactive declaration cycle (`$:`) blocking `vite build`
- Several facility form components use deprecated `on:` event directives (Svelte 4 syntax)
