# Project Structure Proposal

## Goal

You want to keep building the current application, but also create a second website that can reuse:

- UI components
- language/i18n
- theme/tokens
- API client/domain models
- database
- backend modules such as auth, user, team

That is a good reason to move toward a monorepo with shared packages instead of copying code between apps.

## Current Assessment

The project is close, but not fully structured for reuse yet.

What is already good:

- one root `docker-compose.yml` for local full-stack development
- one shared root `.env`
- extracted shared frontend packages now exist for `ui-svelte`, `theme`, and `i18n`
- backend feature modules now exist under `backend/internal/modules` for `auth`, `user`, `team`, and `rbac`
- backend already contains useful generic business areas such as auth, user, team, role, permission
- `apps/web-main` already has a large reusable component base under `src/lib`

What is not good yet:

- `apps/web-main` still contains app-specific code mixed with code that should later move into shared packages
- only the first shared frontend packages have been extracted; the API client/SDK is still app-local
- backend modules are clearer now, but the HTTP handler layer is still mostly centralized
- Compose is fine for one app, but it is not yet shaped for multiple frontends or multiple backend entrypoints

## Recommended Target Structure

Keep one repository and split it by apps and packages.

```text
go_infra_link/
  apps/
    web-main/              # current frontend app
    web-admin/             # second frontend app later
    api-core/              # current backend app
  packages/
    ui-svelte/             # shared Svelte components
    theme/                 # design tokens, Tailwind/CSS config
    i18n/                  # translations and i18n helpers
    typescript-sdk/        # shared API client, DTOs, mappers
  backend/
    modules/
      auth/
      users/
      teams/
      roles/
      permissions/
      projects/
      facility/
    platform/
      db/
      http/
      config/
  infra/
    compose/
      docker-compose.yml
    postgres/
  docs/
```

## Practical Version For Your Stack

Because your backend is Go and your frontend is Svelte, I would not force everything into one package system. Use a mixed monorepo:

- `apps/` for runnable applications
- `packages/` for shared frontend packages
- `backend/modules/` for reusable Go business modules inside the backend codebase

That means:

1. Keep one Go module for now.
2. Keep the current frontend in `apps/web-main`.
3. Add the second frontend later as `apps/web-admin` or `apps/web-second`.
4. Extract shared Svelte code into `packages/ui-svelte`, `packages/theme`, `packages/i18n`, and `packages/typescript-sdk`.
5. Refactor backend feature areas so `auth`, `user`, `team`, and RBAC are explicit modules with clear interfaces.

## Backend Direction

For the backend, do not create a second unrelated backend by copying the current one.

Use one of these two options:

### Option A: One backend, multiple frontend apps

Use this if both websites share:

- users
- teams
- auth
- the same database
- mostly the same business domain

Structure:

```text
backend/
  cmd/
    api/
  internal/
    modules/
      auth/
      user/
      team/
      rbac/
      project/
      facility/
    platform/
      config/
      db/
      http/
```

This is the best option for your current situation.

### Option B: Separate backends with a shared core

Use this only if the second product will become operationally independent.

Structure:

```text
backend/
  services/
    core-api/
    second-api/
  internal/
    shared/
      auth/
      user/
      team/
      rbac/
```

This adds complexity early, so I would not start here.

## Frontend Direction

Right now `apps/web-main` behaves like an app, but the configuration still looks like a Svelte library template.

You should split it into:

- one app package for the current website
- one future app package for the second website
- shared packages for reusable UI and client logic

Recommended shape:

```text
apps/
  web-main/
    src/routes/
    src/app.html
  web-second/
    src/routes/
packages/
  ui-svelte/
    src/lib/components/
  theme/
    src/
  i18n/
    src/
  typescript-sdk/
    src/
```

## Docker Compose Direction

Your current Compose setup is usable for local development, but for future growth it should become more explicit.

Recommended direction:

- keep one root `.env`
- keep one `postgres` service
- keep one `api` service
- add one service per frontend app
- add dev overrides only for local bind mounts and hot reload

Example target:

```yaml
services:
  api:
  web-main:
  web-second:
  postgres:
  pgadmin:
```

Also clean up env usage:

- avoid passing the entire `.env` file into every service unless needed
- keep service-specific env minimal
- keep secrets and local defaults separate if you later deploy outside local Docker

## Suggested Migration Path

Do this in order:

1. Clean the current repo without changing behavior.
2. Continue shrinking `apps/web-main` by extracting more reusable frontend parts into packages.
3. Keep moving backend module boundaries upward from service/domain/repository into routing/handler composition.
4. Extract the shared TypeScript API/client SDK.
5. Add the second frontend app.
6. Extend Compose for two frontend services.

## Concrete Next Tasks

If you want to continue safely, I would do these tasks next:

1. Keep `apps/web-main` as the main web app and remove leftover library-template setup.
2. Add a root workspace layout for frontend packages and apps.
3. Extract shared UI components from the current frontend into a reusable package.
4. Extract shared i18n and theme config.
5. Refactor backend folders so shared domains like auth/team/user are explicit modules.
6. Update Compose to name services by role: `api`, `web-main`, `postgres`, `pgadmin`.

## Recommendation

For your case, the best near-term structure is:

- one repository
- one database
- one backend API
- two frontend apps
- shared frontend packages
- backend feature modules for auth/user/team/RBAC

That gives you reuse without the complexity of maintaining two separate backend systems too early.
