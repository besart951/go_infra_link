# build stage
FROM golang:1.26-alpine AS base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# builder stage
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/app

# runtime
FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /src/translations /app/translations

RUN adduser -D -u 1000 appuser && chown -R 1000:1000 /app
ENV GIN_MODE=release
EXPOSE 8080
USER 1000:1000
ENTRYPOINT ["/app/server"]

# development stage
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
CMD ["air", "--build.cmd", "go build -o tmp/main ./cmd/app", "--build.entrypoint", "./tmp/main"]
