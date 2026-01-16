# build stage
FROM golang:1.25 AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cmd/server ./cmd/server

# runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/cmd/server /app/server
COPY configs/.example.env /app/.env
ENV GIN_MODE=release
EXPOSE 8080
USER 1000:1000
ENTRYPOINT ["/app/server"]
