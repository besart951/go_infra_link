# go_infra_link

Full-stack infrastructure link project with Go backend and future frontend.

## Quick Start

### Backend

```bash
cd backend
go run ./cmd/db-bootstrap
go run ./cmd/app
```

## Development

### Prerequisites

- Go 1.25.6 or higher
- Docker (optional, for containerized deployment)

### Running with Docker

```bash
docker compose up --build
```

### Running Tests

```bash
cd backend
go test ./...
```

## Swagger

### Generate Swagger docs

From the backend folder:

```bash
cd backend
swag init -g ./cmd/app/main.go -o ./docs
```

### View Swagger UI

If `SWAGGER_ENABLED=true`, start the backend and open:

```
http://localhost:8080/swagger/index.html
```

## Architecture

This project follows **Clean/Hexagonal Architecture** principles:

- **Backend**: Go service with layered architecture (domain, application, infrastructure)
- **Frontend**: Separate frontend application (to be implemented)

## Utility Commands

Get a CSRF token from a local login:

```powershell
Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/v1/auth/login" -ContentType "application/json" -Body '{"email":"besart_morina@hotmail.com","password":"password"}' -SessionVariable s | Select-Object -ExpandProperty csrf_token
```

Seed snapshot data into the database:

```bash
docker compose exec backend go run ./cmd/seeder
```

## License

See [LICENSE](LICENSE) for details.
