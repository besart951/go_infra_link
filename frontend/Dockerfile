# Base Stage
FROM node:22-alpine AS base
WORKDIR /app
# Enable pnpm
RUN npm install -g pnpm
COPY pnpm-lock.yaml package.json ./
RUN pnpm install --frozen-lockfile

# Development Stage
FROM base AS dev
ENV CI=true
CMD ["/bin/sh", "-c", "pnpm install && pnpm dev --host"]

# Frontend Build Stage
FROM base AS builder
# Copy source
COPY . .

# Create missing config files from examples
RUN cp src/lib/server/backend.example.ts src/lib/server/backend.ts || true
RUN cp src/lib/server/set-cookie.example.ts src/lib/server/set-cookie.ts || true

# Build
RUN pnpm build

# Serve Stage
FROM caddy:alpine

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/build /usr/share/caddy

EXPOSE 80
