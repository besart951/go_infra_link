# Base Stage
FROM node:25-alpine AS base
WORKDIR /app
# Enable pnpm
RUN npm install -g pnpm
COPY pnpm-lock.yaml package.json ./
RUN pnpm install --frozen-lockfile --ignore-scripts

# Frontend Build Stage
FROM base AS builder
COPY . .

# Build
RUN pnpm build

# Serve Stage
FROM caddy:2.11 AS runtime

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/build /usr/share/caddy

EXPOSE 80

# Development Stage
FROM base AS dev
ENV CI=true
CMD ["/bin/sh", "-c", "pnpm install && pnpm dev --host --port ${FRONTEND_INTERNAL_PORT:-5173}"]
