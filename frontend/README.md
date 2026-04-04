# go_infra_link frontend

SvelteKit frontend for `go_infra_link`.

## Runtime model

This frontend is deployed as a static SPA.

- `@sveltejs/adapter-static` builds static assets into `build/`
- the frontend container serves those files via Caddy
- the edge Caddy instance keeps `/api/*` on the same origin by reverse-proxying to the Go backend
- authentication, sessions, authorization, CSRF, SSE, and all business APIs are owned by the backend

There is no SvelteKit server runtime in production. Do not add `hooks.server.ts`, `+server.ts`, or server-only environment dependencies unless the deployment model is explicitly changed to a server adapter.

## Development

Install dependencies and start the dev server:

```sh
pnpm install
pnpm dev
```

The Vite dev server proxies `/api/*` to the backend so local development matches the production same-origin contract as closely as possible.

## Checks

```sh
pnpm check
pnpm test
pnpm build
```

## Production build

```sh
pnpm build
docker build --target runtime -t local/go-infra-frontend:stable .
```

The runtime image contains only static files and a small Caddy configuration to serve the SPA fallback.
