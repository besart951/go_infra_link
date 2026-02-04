# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Full-stack infrastructure management application. Go backend (Gin + GORM) with SvelteKit frontend (Svelte 5 + Tailwind CSS + bits-ui).

## Commands

### Backend (`cd backend`)

| Task | Command |
|------|---------|
| Run server | `go run ./cmd/app` |
| Build | `go build ./cmd/app` |
| Lint | `make lint` |
| Lint + fix | `make lint-fix` |
| Migrate up | `make migrate-up` |
| Migrate down | `make migrate-down` |
| Swagger docs | `swag init -g ./cmd/app/main.go -o ./docs` |
| Seed data | `go run ./cmd/seeder` |

### Frontend (`cd frontend`)

| Task | Command |
|------|---------|
| Dev server | `pnpm run dev` |
| Build | `pnpm run build` |
| Type check | `pnpm run check` |
| Type check (watch) | `pnpm run check:watch` |
| Format | `pnpm run format` |
| Lint (Prettier) | `pnpm run lint` |

### Full Stack

`docker-compose up` from root (backend + PostgreSQL).

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
  db/                    → Database connection (supports SQLite, PostgreSQL, MySQL)
```

**Key patterns:**
- REST API at `/api/v1/*`, Swagger at `/swagger/index.html`
- JWT access tokens (15min default) + refresh tokens in cookies
- CSRF double-submit cookie pattern for state-changing requests
- GORM AutoMigrate runs at startup (no separate migration step needed for dev)
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

- Backend config via `.env` (see `backend/example.env`): `DB_TYPE`, `DATABASE_URL`, `JWT_SECRET`, `SEED_USER_*`
- Default dev DB: SQLite at `backend/data/app.db`
- Frontend dev server proxies to backend at `localhost:8080`

## Pre-existing Issues

The following build/check issues pre-date current work and are not regressions:
- `svelte-check` reports module resolution errors for non-`.js` imports (e.g., `$lib/api/users` instead of `$lib/api/users.js`)
- `FieldDeviceForm.svelte` has a Svelte 4 reactive declaration cycle (`$:`) blocking `vite build`
- Several facility form components use deprecated `on:` event directives (Svelte 4 syntax)
