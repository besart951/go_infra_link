# Versioned Migrations

This directory contains SQL migration files for production use.

## Structure

Migration files follow the naming convention:

```
{version}_{description}.up.sql   - Forward migration
{version}_{description}.down.sql - Rollback migration
```

Example:

```
000001_create_users_table.up.sql
000001_create_users_table.down.sql
```

## Usage

### Running Migrations

Recommended (no external tooling required):

```bash
# from backend/
go run ./cmd/migrate -path ./migrations -database "$DATABASE_URL" up
go run ./cmd/migrate -path ./migrations -database "$DATABASE_URL" down 1
go run ./cmd/migrate -path ./migrations -database "$DATABASE_URL" version
```

Using golang-migrate CLI:

```bash
# Apply all migrations
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback one migration
migrate -path ./migrations -database "postgres://..." down 1

# Go to specific version
migrate -path ./migrations -database "postgres://..." goto 3
```

### Creating New Migrations

```bash
migrate create -ext sql -dir ./migrations -seq create_projects_table
```

## Development vs Production

- **Development**: apply versioned SQL migrations
- **Production**: apply versioned SQL migrations

Schema changes should be done via new migration files (SQL is the source of truth).
