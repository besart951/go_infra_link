# web-main

Primary Svelte frontend application for the repository.

## Develop

From the repo root:

```sh
docker compose up --build
```

Or run only the web app locally:

```sh
cd apps/web-main
pnpm install
pnpm dev
```

Local runtime note:

- the current Vite/Svelte toolchain requires Node `^20.19.0 || ^22.12.0 || >=24.0.0`
- if your local Node version is older, use Docker Compose instead

## Build

```sh
cd apps/web-main
pnpm build
pnpm preview
```

## Notes

- This folder is an application, not a publishable library package.
- Shared UI, theme, i18n, and API/client code should gradually move into `packages/`.
- A second frontend app can be added alongside this one without duplicating backend ownership.
