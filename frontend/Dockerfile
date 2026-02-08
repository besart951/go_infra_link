# Frontend Build Stage
FROM node:22-alpine AS builder

WORKDIR /app

# Enable pnpm
RUN npm install -g pnpm

# Install dependencies
COPY pnpm-lock.yaml package.json ./
RUN pnpm install --frozen-lockfile

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
