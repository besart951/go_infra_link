# Base stage
FROM node:22-alpine AS base
WORKDIR /workspace
RUN npm install -g pnpm

# Dependency stage for the monorepo workspace
FROM base AS deps
COPY pnpm-lock.yaml package.json pnpm-workspace.yaml ./
COPY apps/web-main/package.json apps/web-main/package.json
COPY packages/ui-svelte/package.json packages/ui-svelte/package.json
COPY packages/theme/package.json packages/theme/package.json
COPY packages/i18n/package.json packages/i18n/package.json
RUN pnpm install --frozen-lockfile

# Development stage
FROM deps AS dev
ENV CI=true
CMD ["/bin/sh", "-c", "pnpm install --frozen-lockfile && pnpm --dir apps/web-main dev --host 0.0.0.0 --port ${FRONTEND_INTERNAL_PORT:-5173}"]

# Frontend build stage
FROM deps AS builder
COPY apps ./apps
COPY packages ./packages
RUN pnpm --dir apps/web-main build

# Serve stage
FROM caddy:2.11
COPY apps/web-main/Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /workspace/apps/web-main/build /usr/share/caddy
EXPOSE 80
