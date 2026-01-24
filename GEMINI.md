# GEMINI.md

## Project Overview

This is a full-stack web application with a Go backend and a SvelteKit frontend. The project is designed with a Clean/Hexagonal Architecture. The backend provides a RESTful API, and the frontend consumes it. The project also includes a `docker-compose.yml` file for easy setup and deployment.

**Backend:**

*   **Language:** Go
*   **Framework:** Gin
*   **Database:** PostgreSQL (with GORM)
*   **Dependencies:** `jwt-go`, `godotenv`, `uuid`, etc.
*   **Testing:** Manual API testing with Bruno.

**Frontend:**

*   **Framework:** SvelteKit
*   **Build Tool:** Vite
*   **Styling:** Tailwind CSS
*   **Type Checking:** Svelte-check
*   **Formatting:** Prettier

## Building and Running

### With Docker (recommended)

The easiest way to run the application is with Docker Compose:

```bash
docker-compose up -d
```

This will start the backend and a PostgreSQL database. The backend will be available at `http://localhost:8080`.

### Backend (Standalone)

1.  **Install dependencies:**
    ```bash
    cd backend
    go mod tidy
    ```
2.  **Run the application:**
    ```bash
    go run ./cmd/app/main.go
    ```
    The backend will be available at `http://localhost:8080`.

### Frontend (Standalone)

1.  **Install dependencies:**
    ```bash
    cd frontend
    pnpm install
    ```
2.  **Run the development server:**
    ```bash
    pnpm dev
    ```
    The frontend will be available at `http://localhost:5173`.

## Development Conventions

### Backend

*   **Linting:** The project uses `golangci-lint` for linting. You can run the linter with `make lint` and fix issues with `make lint-fix` in the `backend` directory.
*   **Database Migrations:** Database migrations are managed with `go-migrate`. Use the `make` commands in the `backend` directory to run migrations.
*   **Configuration:** Backend configuration is managed through environment variables and a `.env` file. See `backend/example.env` for a list of available variables.

### Frontend

*   **Formatting:** The project uses Prettier for code formatting. You can format the code with `pnpm format` in the `frontend` directory.
*   **Type Checking:** The project uses `svelte-check` for type checking. You can run the type checker with `pnpm check` in the `frontend` directory.
*   **API Client:** A Bruno collection is provided in the `bruno` directory for manual API testing.

## Key Files

*   `backend/`: The Go backend.
*   `frontend/`: The SvelteKit frontend.
*   `docker-compose.yml`: Docker Compose file for running the application.
*   `bruno/`: Bruno collection for API testing.
*   `ARCHITECTURE.md`: Detailed documentation of the project's architecture.
*   `README.md`: The original README file.
