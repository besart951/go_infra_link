# build stage
FROM golang:1.26-alpine AS base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# builder stage
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/db-bootstrap ./cmd/db-bootstrap

# bootstrap runtime — pre-compiled DB migration binary (used in production)
FROM alpine:3.23 AS bootstrap
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/db-bootstrap /app/db-bootstrap
RUN adduser -D -u 1000 appuser && chown -R 1000:1000 /app
USER 1000:1000
ENTRYPOINT ["/app/db-bootstrap"]

# runtime
FROM alpine:3.23 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /src/translations /app/translations
RUN adduser -D -u 1000 appuser && chown -R 1000:1000 /app
ENV GIN_MODE=release
EXPOSE 8080
USER 1000:1000
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/server"]

# development stage
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
CMD ["air", "--build.cmd", "go build -o tmp/main ./cmd/app", "--build.entrypoint", "./tmp/main"]
