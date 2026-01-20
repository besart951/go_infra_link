# API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Overview
This API provides CRUD operations for Projects and Users, following the Go-Way architecture pattern with proper separation of concerns between routers, handlers, and services.

---

## Projects API

### 1. Create Project
**Endpoint:** `POST /api/v1/projects`

**Request Body:**
```json
{
  "name": "New Project",
  "description": "Project description",
  "status": "planned",
  "start_date": "2024-01-15T10:00:00Z",
  "phase_id": "550e8400-e29b-41d4-a716-446655440000",
  "creator_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Validation Rules:**
- `name`: Required, 1-255 characters
- `description`: Optional
- `status`: Optional, one of: `planned`, `ongoing`, `completed` (default: `planned`)
- `start_date`: Optional, ISO 8601 timestamp
- `phase_id`: Optional, UUID format
- `creator_id`: Required, UUID format

**Success Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "New Project",
  "description": "Project description",
  "status": "planned",
  "start_date": "2024-01-15T10:00:00Z",
  "phase_id": "550e8400-e29b-41d4-a716-446655440000",
  "creator_id": "550e8400-e29b-41d4-a716-446655440001",
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-15T09:00:00Z"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "validation_error",
  "message": "Key: 'CreateProjectRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

---

### 2. Get Project by ID
**Endpoint:** `GET /api/v1/projects/:id`

**URL Parameters:**
- `id`: UUID of the project

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "New Project",
  "description": "Project description",
  "status": "planned",
  "start_date": "2024-01-15T10:00:00Z",
  "phase_id": "550e8400-e29b-41d4-a716-446655440000",
  "creator_id": "550e8400-e29b-41d4-a716-446655440001",
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-15T09:00:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "not_found",
  "message": "Project not found"
}
```

---

### 3. List Projects
**Endpoint:** `GET /api/v1/projects`

**Query Parameters:**
- `page`: Page number (optional, default: 1, min: 1)
- `limit`: Items per page (optional, default: 10, min: 1, max: 100)
- `search`: Search query to filter by name or description (optional)

**Example:** `GET /api/v1/projects?page=1&limit=10&search=infrastructure`

**Success Response (200 OK):**
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "Infrastructure Project",
      "description": "Main infrastructure project",
      "status": "ongoing",
      "start_date": "2024-01-15T10:00:00Z",
      "phase_id": "550e8400-e29b-41d4-a716-446655440000",
      "creator_id": "550e8400-e29b-41d4-a716-446655440001",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-16T14:30:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "total_pages": 3
}
```

---

### 4. Update Project
**Endpoint:** `PUT /api/v1/projects/:id`

**URL Parameters:**
- `id`: UUID of the project

**Request Body:**
```json
{
  "name": "Updated Project Name",
  "description": "Updated description",
  "status": "ongoing",
  "start_date": "2024-01-20T10:00:00Z",
  "phase_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

**Validation Rules:**
- All fields are optional
- `name`: 1-255 characters if provided
- `status`: Must be one of: `planned`, `ongoing`, `completed` if provided
- `phase_id`: UUID format if provided

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "Updated Project Name",
  "description": "Updated description",
  "status": "ongoing",
  "start_date": "2024-01-20T10:00:00Z",
  "phase_id": "550e8400-e29b-41d4-a716-446655440003",
  "creator_id": "550e8400-e29b-41d4-a716-446655440001",
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-20T11:15:00Z"
}
```

---

### 5. Delete Project
**Endpoint:** `DELETE /api/v1/projects/:id`

**URL Parameters:**
- `id`: UUID of the project

**Success Response (204 No Content):**
- Empty response body

**Error Response (400 Bad Request):**
```json
{
  "error": "invalid_id",
  "message": "Invalid UUID format"
}
```

---

## Users API

### 1. Create User
**Endpoint:** `POST /api/v1/users`

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "password": "securePassword123",
  "is_active": true,
  "created_by_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Validation Rules:**
- `first_name`: Required, 1-100 characters
- `last_name`: Required, 1-100 characters
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters
- `is_active`: Optional, boolean (default: false)
- `created_by_id`: Optional, UUID format

**Success Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "is_active": true,
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-15T09:00:00Z"
}
```

**Note:** Password is not returned in responses for security reasons.

---

### 2. Get User by ID
**Endpoint:** `GET /api/v1/users/:id`

**URL Parameters:**
- `id`: UUID of the user

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "is_active": true,
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-15T09:00:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "not_found",
  "message": "User not found"
}
```

---

### 3. List Users
**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page`: Page number (optional, default: 1, min: 1)
- `limit`: Items per page (optional, default: 10, min: 1, max: 100)
- `search`: Search query to filter by first name, last name, or email (optional)

**Example:** `GET /api/v1/users?page=1&limit=20&search=john`

**Success Response (200 OK):**
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "is_active": true,
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T09:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "total_pages": 3
}
```

---

### 4. Update User
**Endpoint:** `PUT /api/v1/users/:id`

**URL Parameters:**
- `id`: UUID of the user

**Request Body:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "password": "newSecurePassword456",
  "is_active": false
}
```

**Validation Rules:**
- All fields are optional
- `first_name`: 1-100 characters if provided
- `last_name`: 1-100 characters if provided
- `email`: Valid email format if provided
- `password`: Minimum 8 characters if provided
- `is_active`: Boolean if provided

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "is_active": false,
  "created_at": "2024-01-15T09:00:00Z",
  "updated_at": "2024-01-20T14:30:00Z"
}
```

---

### 5. Delete User
**Endpoint:** `DELETE /api/v1/users/:id`

**URL Parameters:**
- `id`: UUID of the user

**Success Response (204 No Content):**
- Empty response body

**Error Response (400 Bad Request):**
```json
{
  "error": "invalid_id",
  "message": "Invalid UUID format"
}
```

---

## Health Check

### Health Check Endpoint
**Endpoint:** `GET /health`

**Success Response (200 OK):**
```json
{
  "status": "ok"
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

### Common Error Codes:
- `validation_error` (400): Request validation failed
- `invalid_id` (400): Invalid UUID format
- `not_found` (404): Resource not found
- `creation_failed` (500): Failed to create resource
- `update_failed` (500): Failed to update resource
- `deletion_failed` (500): Failed to delete resource
- `fetch_failed` (500): Failed to fetch resource

---

## Architecture

### Go-Way Pattern Implementation

The API follows the Go-Way architecture pattern with clear separation of concerns:

1. **Domain Layer** (`internal/domain/`)
   - Defines business entities (User, Project)
   - Defines repository interfaces
   - Contains business logic independent of infrastructure

2. **Repository Layer** (`internal/repository/`)
   - Implements data access using GORM
   - Handles database operations
   - Provides pagination and search capabilities

3. **Service Layer** (`internal/service/`)
   - Contains business logic
   - Orchestrates repository operations
   - Provides a clean API for handlers

4. **Handler Layer** (`internal/handler/`)
   - HTTP request/response handling
   - Input validation using struct tags
   - DTOs for request/response transformation
   - Error handling and status codes

5. **Router Layer** (`internal/handler/routes.go`)
   - Route registration
   - Endpoint grouping
   - Middleware integration point

### Key Features:
- **Validation**: Automatic request validation using Gin's binding tags
- **Pagination**: Built-in pagination support for list endpoints
- **Search**: Full-text search across relevant fields
- **UUID v7**: Modern UUID generation with timestamp ordering
- **Soft Deletes**: GORM soft delete support
- **Error Handling**: Consistent error response format
- **Type Safety**: Strong typing with Go generics for repositories

---

## Notes

1. **Password Security**: The current implementation stores passwords in plain text. In production, passwords should be hashed using bcrypt or similar before storage.

2. **Authentication**: No authentication is currently implemented. In production, add JWT or session-based authentication.

3. **Database**: The API uses GORM with support for PostgreSQL and SQLite. Connection is configured via environment variables.

4. **Validation**: All validation is performed using go-playground/validator via Gin's binding tags.

5. **UUIDs**: All IDs are UUID v7 format, which includes timestamp ordering for better database performance.
