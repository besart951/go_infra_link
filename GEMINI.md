# GEMINI.md

## Project Overview

This is a robust, full-stack web application designed for scalability and maintainability. It connects a high-performance **Go backend** with a modern, reactive **SvelteKit frontend**.

The project is built upon strong software engineering foundations, strictly adhering to **SOLID principles** and **Clean/Hexagonal Architecture** to ensure loose coupling and high testability.

### Tech Stack

**Backend:**

- **Language:** Go (1.25.6)
- **Framework:** Gin (Web Framework)
- **Database:** PostgreSQL or MySQL/MariaDB
- **ORM:** GORM (with AutoMigrate enabled)
- **Dependency Injection:** `google/wire` (Compile-time DI)
- **Configuration:** `godotenv` & environment variables
- **Documentation:** Swagger/OpenAPI (Auto-generated)
- **Testing:** Manual API testing via Bruno collections

**Frontend:**

- **Framework:** SvelteKit (Svelte 5)
- **Build Tool:** Vite 7
- **Styling:** Tailwind CSS 4
- **UI Library:** `bits-ui` (headless components), `lucide-svelte` (icons)
- **Language:** TypeScript
- **Quality:** Prettier, ESLint, Svelte-check

## Architecture & Design Principles

### Backend: The "Go Way" & Hexagonal Architecture

The backend is structured to keep business logic independent of external concerns (frameworks, databases, UI).

- **Hexagonal Architecture (Ports & Adapters):**
  - **`internal/domain`**: The core. Contains entities, repository interfaces, and domain errors. purely Go structs and interfaces. No external dependencies.
  - **`internal/service`**: Application business logic. Implements use cases using domain repositories.
  - **`internal/repository`**: Adapters that implement domain interfaces (e.g., GORM implementations of repositories).
  - **`internal/handler`**: HTTP adapters (Gin handlers) that map HTTP requests to service calls and responses.
  - **`internal/app`**: The composition root. Wires everything together using `wire`.

- **SOLID Principles:**
  - **Single Responsibility:** Each package and struct has a distinct purpose (e.g., `user_handler` only handles HTTP, `user_service` handles logic).
  - **Dependency Inversion:** High-level modules (Services) depend on abstractions (Repository Interfaces), not details (GORM implementations).
  - **Interface Segregation:** Interfaces are defined where they are used (in the domain or service layer), keeping them small and focused.

- **Idiomatic Go:**
  - Error handling is explicit (`if err != nil`).
  - Context is propagated through all layers for cancellation and timeouts.
  - Configuration is explicit and typed.

### Frontend: Centralized Fetching & Modern UX

The frontend is designed for developer productivity and a consistent user experience.

- **Central Fetching Strategy (`src/lib/api/client.ts`):**
  - A unified `api()` wrapper around `fetch`.
  - **Automatic CSRF Handling:** Extracts and attaches `X-CSRF-Token` headers.
  - **Error Normalization:** Converts all backend errors into a standard `ApiException` or `HandledApiException`.
  - **Global Error Handling:** Automatically triggers UI feedback (e.g., toast notifications) for common issues like authorization failures (`403 Forbidden`).
  - **Type Safety:** Generic support for strongly-typed API responses.

- **Component Architecture:**
  - **Headless UI:** Uses `bits-ui` for accessible, unstyled primitives, styled via Tailwind.
  - **Atomic Design:** Reusable components live in `src/lib/components`.

## Development Workflows

### 1. Code Quality & Formatting (CRITICAL)

Strict formatting and linting are enforced to maintain a consistent codebase.

**Backend (Go):**

- **Formatting:** Standard `gofmt` is mandatory.
- **Linting:** `golangci-lint` is the source of truth.
  - Run: `make lint`
  - Fix: `make lint-fix`

**Frontend (TypeScript/Svelte):**

- **Formatting:** Prettier is mandatory.
  - Run: `pnpm format` (Write changes)
  - Check: `pnpm lint` (Check only)
- **Type Checking:** Svelte-check ensures type safety across `.svelte` and `.ts` files.
  - Run: `pnpm check`

### 2. Database Management

The backend uses **GORM AutoMigrate**.

- **Workflow:** Simply define your struct in `internal/domain` and register it in `internal/db/database.go`.
- **Startup:** The application automatically synchronizes the database schema with your Go structs when it starts. No manual SQL migration files are strictly required for standard operations.

### 3. API Documentation

Swagger documentation is automatically generated from code comments.

- **Generate:** `swag init -g cmd/app/main.go -o docs` (or via Makefile if configured).
- **View:** `http://localhost:8080/swagger/index.html` (in Development mode).

## Setup & Run

### Docker (Recommended)

Spins up the entire stack (Backend + Frontend + DB).

```bash
docker-compose up -d
```

- **Backend:** `http://localhost:8080`
- **Frontend:** `http://localhost:5173`

### Manual Setup

**Backend:**

```bash
cd backend
go mod tidy
go run ./cmd/app/main.go
```

**Frontend:**

```bash
cd frontend
pnpm install
pnpm dev
```

## Key Directory Structure

```text
├── backend/
│   ├── cmd/            # Entry points (app, seeder)
│   ├── internal/
│   │   ├── domain/     # Entities & Interfaces (Pure Go)
│   │   ├── service/    # Business Logic
│   │   ├── repository/ # Database Implementations
│   │   ├── handler/    # HTTP Controllers
│   │   └── wire/       # Dependency Injection Setup
│   └── makefile        # Linting & helper commands
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/    # Centralized Fetch Client
│   │   │   └── ...
│   │   └── routes/     # SvelteKit Pages
│   └── package.json    # Scripts (format, check, etc.)
└── bruno/              # API Testing Collection
```
