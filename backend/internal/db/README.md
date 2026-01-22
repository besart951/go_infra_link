# Database Package

This package provides database connectivity and migration using GORM.

## Features

- **Environment-driven Configuration**: Switch between SQLite, PostgreSQL, and MySQL/MariaDB using the `DB_TYPE` environment variable
- **Auto-Migration**: GORM's AutoMigrate automatically creates and updates database schema from Go structs
- **Connection Pooling**: Configurable connection pool settings for optimal performance
- **Clean Architecture**: Minimal boilerplate with clear separation of concerns

## Supported Databases

### SQLite
```env
DB_TYPE=sqlite
DB_DSN=./data/app.db
```

### PostgreSQL
```env
DB_TYPE=postgres
DB_DSN=host=localhost port=5432 user=postgres password=postgres dbname=go_infra_link sslmode=disable
```

### MySQL/MariaDB
```env
DB_TYPE=mysql
DB_DSN=user:password@tcp(localhost:3306)/go_infra_link?charset=utf8mb4&parseTime=True&loc=Local
```

## Usage

```go
import "github.com/besart951/go_infra_link/backend/internal/db"

// Connect to database and run migrations
gormDB, err := db.Connect(cfg)
if err != nil {
    return fmt.Errorf("db connect: %w", err)
}

// Get underlying sql.DB for existing repositories
sqlDB, err := gormDB.DB()
```

## Migration

All domain models are automatically migrated when the application starts. Models are defined in the `internal/domain` package with GORM tags.

No manual SQL migration files are needed - GORM handles schema creation and updates automatically.

## Models

All models are listed in the `autoMigrate()` function:
- User & BusinessDetails
- RefreshToken, PasswordResetToken, LoginAttempt
- Team & TeamMember
- Phase & Project
- Facility models (Building, ControlCabinet, SPSController, etc.)
