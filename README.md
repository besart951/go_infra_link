# go_infra_link

Full-stack infrastructure link project with Go backend and future frontend.

## Quick Start

### Backend

```bash
cd backend
go run ./cmd/server
```

## Development

### Prerequisites

- Go 1.25.6 or higher
- Docker (optional, for containerized deployment)

### Running with Docker

```bash
docker build -t go_infra_link .
docker run -p 8080:8080 go_infra_link
```

### Using Make

```bash
# Build backend
make build

# Run backend
make run

# Run tests
make test

# Clean build artifacts
make clean
```

## Swagger

### Generate Swagger docs

From the backend folder:

```bash
cd backend
swag init -g ./cmd/app/main.go -o ./docs
```

### View Swagger UI

Start the backend and open:

```
http://localhost:8080/swagger/index.html
```

## Architecture

This project follows **Clean/Hexagonal Architecture** principles:

- **Backend**: Go service with layered architecture (domain, application, infrastructure)
- **Frontend**: Separate frontend application (to be implemented)

## License

See [LICENSE](LICENSE) for details.

CSRF-Token
Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/v1/auth/login" -ContentType "application/json" -Body '{"email":"besart_morina@hotmail.com","password":"password"}' -SessionVariable s | Select-Object -ExpandProperty csrf_token


docker compose exec backend go run ./cmd/seeder
