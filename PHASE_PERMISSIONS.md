# Phase-Based Role Permission System

This document describes the phase-based role permission system implemented in the application.

## Overview

The system allows fine-grained access control where user permissions can vary by phase. For example:
- In phase "SIA:51", users with `RoleAdminPlaner` can **edit**
- In phase "SIA:61", users with `RoleAdminPlaner` can only **suggest_changes**

## Key Features

1. **Auto-generated Phase IDs with UUID7**: All phases automatically get a UUID v7 ID when created
2. **Phases are per project**: Each phase belongs to a specific project
3. **Role-based permissions per phase**: Different roles can have different permissions in different phases
4. **Flexible permission types**: Support for edit, suggest_changes, view, delete, and manage_users

## Database Schema

### Phases Table
```sql
CREATE TABLE `phases` (
  `id` uuid PRIMARY KEY,
  `created_at` datetime,
  `updated_at` datetime,
  `deleted_at` datetime,
  `name` text NOT NULL,
  `project_id` uuid NOT NULL,
  FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`)
);
```

### Phase Permissions Table
```sql
CREATE TABLE `phase_permissions` (
  `id` uuid PRIMARY KEY,
  `created_at` datetime,
  `updated_at` datetime,
  `deleted_at` datetime,
  `phase_id` uuid NOT NULL,
  `role` varchar(50) NOT NULL,
  `permission` varchar(50) NOT NULL,
  FOREIGN KEY (`phase_id`) REFERENCES `phases`(`id`),
  UNIQUE (`phase_id`, `role`)
);
```

## API Endpoints

### Phase Management

- `POST /api/v1/phases` - Create a new phase
- `GET /api/v1/phases` - List all phases (with pagination)
- `GET /api/v1/phases/:id` - Get a specific phase
- `PUT /api/v1/phases/:id` - Update a phase
- `DELETE /api/v1/phases/:id` - Delete a phase

### Phase Permission Management

- `POST /api/v1/phase-permissions` - Create a new phase permission
- `GET /api/v1/phase-permissions?phase_id={id}` - List permissions for a phase
- `GET /api/v1/phase-permissions/:id` - Get a specific permission
- `PUT /api/v1/phase-permissions/:id` - Update a permission
- `DELETE /api/v1/phase-permissions/:id` - Delete a permission

## Available Roles

- `user` - Regular user
- `admin` - Administrator
- `superadmin` - Super administrator
- `admin_planer` - Administrator planner
- `planer` - Planner
- `admin_entrepreneur` - Administrator entrepreneur
- `entrepreneur` - Entrepreneur

## Available Permissions

- `edit` - Full edit access
- `suggest_changes` - Can only suggest changes (not directly edit)
- `view` - Read-only access
- `delete` - Can delete items
- `manage_users` - Can manage user access

## Usage Examples

### 1. Create a Project Phase

```json
POST /api/v1/phases
{
  "name": "SIA:51",
  "project_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Response:
```json
{
  "id": "01936d6f-7c8a-7000-8000-000000000000",
  "name": "SIA:51",
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2026-01-28T16:30:00Z",
  "updated_at": "2026-01-28T16:30:00Z"
}
```

### 2. Set Permission for a Role in a Phase

Allow `admin_planer` to edit in phase "SIA:51":

```json
POST /api/v1/phase-permissions
{
  "phase_id": "01936d6f-7c8a-7000-8000-000000000000",
  "role": "admin_planer",
  "permission": "edit"
}
```

### 3. Create Another Phase with Different Permissions

```json
POST /api/v1/phases
{
  "name": "SIA:61",
  "project_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Then set permission for `admin_planer` to only suggest changes:

```json
POST /api/v1/phase-permissions
{
  "phase_id": "01936d6f-7c8a-7000-8001-000000000000",
  "role": "admin_planer",
  "permission": "suggest_changes"
}
```

### 4. List Permissions for a Phase

```
GET /api/v1/phase-permissions?phase_id=01936d6f-7c8a-7000-8000-000000000000
```

Response:
```json
{
  "items": [
    {
      "id": "01936d6f-7c8a-7001-8000-000000000000",
      "phase_id": "01936d6f-7c8a-7000-8000-000000000000",
      "role": "admin_planer",
      "permission": "edit",
      "created_at": "2026-01-28T16:30:00Z",
      "updated_at": "2026-01-28T16:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "total_pages": 1
}
```

## Implementation Details

### Domain Layer

- **Phase**: Represents a project phase with UUID7 ID, name, and project reference
- **PhasePermission**: Maps a role to a permission type for a specific phase
- **PermissionType**: Enum defining available permission types

### Repository Layer

- **PhaseRepository**: CRUD operations for phases
- **PhasePermissionRepository**: CRUD operations for phase permissions with special methods:
  - `GetByPhaseAndRole`: Get permission for a specific phase and role
  - `ListByPhase`: Get all permissions for a phase
  - `DeleteByPhaseAndRole`: Delete a specific phase-role permission

### Service Layer

- **PhaseService**: Business logic for phase management
- **PhasePermissionService**: Business logic for permission management

### Handler Layer

- **PhaseHandler**: HTTP handlers for phase endpoints
- **PhasePermissionHandler**: HTTP handlers for permission endpoints

## Testing

Bruno API collection files are provided in the `bruno/` directory:
- `bruno/phases/` - Phase CRUD operations
- `bruno/phase-permissions/` - Phase permission CRUD operations

## Future Enhancements

1. Add middleware to enforce phase-based permissions in API endpoints
2. Add validation to ensure users can only perform actions they have permission for
3. Add audit logging for permission changes
4. Add bulk permission operations for efficiency
5. Add permission inheritance or templates
