# syntax=docker/dockerfile:1.7

# Base Stage
FROM node:25-alpine AS base
WORKDIR /app
ENV PNPM_STORE_DIR=/pnpm/store
# Enable the same pinned pnpm version used in CI.
RUN npm install -g pnpm@10.29.1
COPY pnpm-lock.yaml package.json ./
RUN --mount=type=cache,target=/pnpm/store pnpm install --frozen-lockfile --ignore-scripts

# Frontend Build Stage
FROM base AS builder
COPY . .

# Build
RUN pnpm build -- --logLevel warn

# Serve Stage
FROM caddy:2.11 AS runtime

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/build /usr/share/caddy

EXPOSE 80

# Development Stage
FROM base AS dev
ENV CI=true
CMD ["/bin/sh", "-c", "pnpm install && pnpm dev --host --port ${FRONTEND_INTERNAL_PORT:-5173}"]
