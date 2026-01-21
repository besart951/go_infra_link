# User Account Management System - API Documentation

This document provides comprehensive API documentation for the User Account Management system with RBAC (Role-Based Access Control).

## Table of Contents
- [Authentication](#authentication)
- [User Management](#user-management)
- [Team Management](#team-management)
- [Admin Operations](#admin-operations)
- [RBAC & Permissions](#rbac--permissions)

---

## Authentication

### Login
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response 200:
{
  "csrf_token": "...",
  "access_token": "...",
  "refresh_token": "..."
}
```

### Refresh Token
```
POST /api/v1/auth/refresh
Cookie: refresh_token=...
X-CSRF-Token: ...

Response 200:
{
  "access_token": "...",
  "csrf_token": "..."
}
```

### Logout
```
POST /api/v1/auth/logout
Cookie: refresh_token=...
X-CSRF-Token: ...

Response 204
```

### Get Current User
```
GET /api/v1/auth/me
Cookie: access_token=...

Response 200:
{
  "id": "uuid",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "role": "admin",
  "is_active": true,
  "company_name": "Acme Corp",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

## User Management

### List Users with Advanced Filtering
**Requires: Admin role**

```
GET /api/v1/users?page=1&limit=10&search=john&role=admin&is_active=true&company_name=Acme
Cookie: access_token=...
X-CSRF-Token: ...

Query Parameters:
- page (int): Page number (default: 1)
- limit (int): Items per page (default: 10, max: 100)
- search (string): Search by first name, last name, or email
- role (string): Filter by role (user, admin, superadmin)
- is_active (string): Filter by active status (true, false)
- company_name (string): Filter by company name

Response 200:
{
  "items": [
    {
      "id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "is_active": true,
      "role": "admin",
      "company_name": "Acme Corp",
      "disabled_at": null,
      "locked_until": null,
      "failed_login_attempts": 0,
      "last_login_at": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "total_pages": 10
}
```

### Get User by ID
**Requires: Admin role**

```
GET /api/v1/users/{id}
Cookie: access_token=...
X-CSRF-Token: ...

Response 200:
{
  "id": "uuid",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "is_active": true,
  "role": "admin",
  "company_name": "Acme Corp",
  "disabled_at": null,
  "locked_until": null,
  "failed_login_attempts": 0,
  "last_login_at": "2024-01-15T10:30:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Create User
**Requires: Admin role**

```
POST /api/v1/users
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "password": "SecurePass123!",
  "is_active": true,
  "role": "user",
  "created_by_id": "admin-uuid"
}

Response 201:
{
  "id": "new-uuid",
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "is_active": true,
  "role": "user",
  "created_at": "2024-01-20T15:00:00Z",
  "updated_at": "2024-01-20T15:00:00Z"
}
```

### Update User
**Requires: Admin role**

```
PUT /api/v1/users/{id}
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Doe",
  "email": "jane.doe@example.com",
  "password": "NewPassword123!",
  "is_active": false,
  "role": "admin"
}

Response 200: (Same as Get User)
```

### Delete User
**Requires: Admin role**

```
DELETE /api/v1/users/{id}
Cookie: access_token=...
X-CSRF-Token: ...

Response 204
```

---

## Team Management

### List Teams
**Requires: Admin role**

```
GET /api/v1/teams?page=1&limit=10&search=engineering
Cookie: access_token=...
X-CSRF-Token: ...

Query Parameters:
- page (int): Page number
- limit (int): Items per page
- search (string): Search by team name or description

Response 200:
{
  "items": [
    {
      "id": "uuid",
      "name": "Engineering Team",
      "description": "Development team",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "total_pages": 1
}
```

### Get Team by ID
**Requires: Team member role or higher**

```
GET /api/v1/teams/{id}
Cookie: access_token=...
X-CSRF-Token: ...

Response 200:
{
  "id": "uuid",
  "name": "Engineering Team",
  "description": "Development team",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Create Team
**Requires: Admin role**

```
POST /api/v1/teams
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "name": "Marketing Team",
  "description": "Marketing and sales"
}

Response 201:
{
  "id": "new-uuid",
  "name": "Marketing Team",
  "description": "Marketing and sales",
  "created_at": "2024-01-20T15:00:00Z",
  "updated_at": "2024-01-20T15:00:00Z"
}
```

### Update Team
**Requires: Team manager role or higher**

```
PUT /api/v1/teams/{id}
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "name": "Updated Team Name",
  "description": "Updated description"
}

Response 200: (Same as Get Team)
```

### Delete Team
**Requires: Team owner role**

```
DELETE /api/v1/teams/{id}
Cookie: access_token=...
X-CSRF-Token: ...

Response 204
```

### Add Team Member
**Requires: Team manager role or higher**

```
POST /api/v1/teams/{id}/members
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "user_id": "user-uuid",
  "role": "member"
}

Roles: member, manager, owner

Response 204
```

### Remove Team Member
**Requires: Team manager role or higher**

```
DELETE /api/v1/teams/{id}/members/{userId}
Cookie: access_token=...
X-CSRF-Token: ...

Response 204
```

### List Team Members
**Requires: Team member role or higher**

```
GET /api/v1/teams/{id}/members?page=1&limit=10
Cookie: access_token=...
X-CSRF-Token: ...

Response 200:
{
  "items": [
    {
      "team_id": "team-uuid",
      "user_id": "user-uuid",
      "role": "manager",
      "joined_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 10,
  "page": 1,
  "total_pages": 1
}
```

---

## Admin Operations

### Reset User Password
**Requires: Admin role**

```
POST /api/v1/admin/users/{id}/password-reset
Cookie: access_token=...
X-CSRF-Token: ...

Response 200:
{
  "reset_token": "temporary-token-123",
  "expires_at": "2024-01-20T16:00:00Z"
}
```

### Disable User
**Requires: Admin role**

```
POST /api/v1/admin/users/{id}/disable
Cookie: access_token=...
X-CSRF-Token: ...

Response 204
```

### Enable User
**Requires: Admin role**

```
POST /api/v1/admin/users/{id}/enable
Cookie: access_token=...
X-CSRF-Token: ...

Response 204
```

### Lock User Account
**Requires: Admin role**

```
POST /api/v1/admin/users/{id}/lock
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "until": "2024-02-01T00:00:00Z"
}

Response 204
```

### Unlock User Account
**Requires: Admin role**

```
POST /api/v1/admin/users/{id}/unlock
Cookie: access_token=...
X-CSRF-Token: ...

Response 204
```

### Set User Role
**Requires: Admin role**

```
POST /api/v1/admin/users/{id}/role
Cookie: access_token=...
X-CSRF-Token: ...
Content-Type: application/json

{
  "role": "admin"
}

Roles: user, admin, superadmin

Response 204
```

### List Login Attempts (Audit Log)
**Requires: Admin role**

```
GET /api/v1/admin/login-attempts?page=1&limit=20&search=john
Cookie: access_token=...
X-CSRF-Token: ...

Response 200:
{
  "items": [
    {
      "id": "uuid",
      "created_at": "2024-01-20T10:00:00Z",
      "user_id": "user-uuid",
      "email": "john.doe@example.com",
      "ip": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "success": true,
      "failure_reason": null
    },
    {
      "id": "uuid",
      "created_at": "2024-01-20T09:55:00Z",
      "user_id": null,
      "email": "invalid@example.com",
      "ip": "192.168.1.2",
      "user_agent": "curl/7.68.0",
      "success": false,
      "failure_reason": "invalid_credentials"
    }
  ],
  "total": 1000,
  "page": 1,
  "total_pages": 50
}
```

---

## RBAC & Permissions

### User Roles (Hierarchical)

1. **superadmin** (Level 100)
   - Full system access
   - Can manage all users and teams
   - Can access all admin operations

2. **admin** (Level 50)
   - Can manage users and teams
   - Can access admin operations
   - Cannot modify superadmin users

3. **user** (Level 10)
   - Standard user access
   - Can only manage their own profile
   - No admin access

### Team Roles (Hierarchical)

1. **owner** (Level 100)
   - Full team control
   - Can delete the team
   - Can manage all members

2. **manager** (Level 50)
   - Can add/remove members
   - Can update team details
   - Cannot delete the team

3. **member** (Level 10)
   - Can view team details
   - Read-only access
   - Cannot modify team

### Permission Matrix

| Action | User | Admin | SuperAdmin | Team Member | Team Manager | Team Owner |
|--------|------|-------|------------|-------------|--------------|------------|
| View own profile | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Update own profile | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| List all users | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Create user | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update any user | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete any user | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Create team | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| View team | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Update team | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Delete team | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ |
| Add team member | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Remove team member | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |
| Reset user password | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Lock/Unlock user | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| View login attempts | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## Security Features

### Password Security
- **Hashing**: Bcrypt with salt
- **Minimum Requirements**: 8 characters (enforced in validation)
- **Password Reset**: Temporary tokens with expiration

### Authentication
- **JWT Tokens**: Stateless authentication
- **Access Tokens**: Short-lived (stored in HTTP-only cookies)
- **Refresh Tokens**: Long-lived (stored in HTTP-only cookies)
- **CSRF Protection**: Required for all state-changing operations

### Account Security
- **Failed Login Tracking**: Automatic tracking of failed attempts
- **Account Locking**: Temporary lock after multiple failed attempts
- **Account Disable**: Manual disable by admins
- **Audit Logging**: All login attempts are logged with IP and user agent

### Request Security
- All authenticated requests require:
  - Valid JWT access token (in cookie)
  - Valid CSRF token (in header for POST/PUT/DELETE)
- Cookies are HTTP-only and Secure (in production)

---

## Frontend Integration Guide

### Setting Up Authentication

1. **Login Flow**:
   ```javascript
   const response = await fetch('/api/v1/auth/login', {
     method: 'POST',
     headers: { 'Content-Type': 'application/json' },
     body: JSON.stringify({ email, password }),
     credentials: 'include' // Important for cookies
   });
   const data = await response.json();
   // Store CSRF token for future requests
   localStorage.setItem('csrf_token', data.csrf_token);
   ```

2. **Making Authenticated Requests**:
   ```javascript
   const csrfToken = localStorage.getItem('csrf_token');
   const response = await fetch('/api/v1/users', {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json',
       'X-CSRF-Token': csrfToken
     },
     body: JSON.stringify(userData),
     credentials: 'include'
   });
   ```

3. **Handling Token Refresh**:
   ```javascript
   // On 401 response with token_expired error
   const refreshResponse = await fetch('/api/v1/auth/refresh', {
     method: 'POST',
     headers: { 'X-CSRF-Token': localStorage.getItem('csrf_token') },
     credentials: 'include'
   });
   const data = await refreshResponse.json();
   localStorage.setItem('csrf_token', data.csrf_token);
   // Retry original request
   ```

### Admin Dashboard Features

#### User Management
- **List View**: Paginated table with filtering by role, status, company
- **Search**: Real-time search by name or email
- **User Details**: Modal/page showing full user info including login stats
- **Actions**: 
  - Create/Edit/Delete users
  - Reset password (generates temporary token)
  - Enable/Disable account
  - Lock/Unlock account
  - Change user role

#### Team Management
- **List View**: All teams with member counts
- **Team Details**: Members list with roles
- **Actions**:
  - Create/Edit/Delete teams
  - Add/Remove members
  - Change member roles

#### Audit & Security
- **Login Attempts**: Searchable log of all login attempts
- **Filters**: Success/failure, date range, user
- **Display**: IP address, user agent, timestamp, result

### Example React Component

```jsx
import React, { useState, useEffect } from 'react';

function UserList() {
  const [users, setUsers] = useState([]);
  const [filters, setFilters] = useState({
    page: 1,
    limit: 10,
    search: '',
    role: '',
    is_active: '',
    company_name: ''
  });
  const [pagination, setPagination] = useState({});

  useEffect(() => {
    fetchUsers();
  }, [filters]);

  const fetchUsers = async () => {
    const params = new URLSearchParams(
      Object.entries(filters).filter(([_, v]) => v !== '')
    );
    
    const response = await fetch(`/api/v1/users?${params}`, {
      credentials: 'include',
      headers: {
        'X-CSRF-Token': localStorage.getItem('csrf_token')
      }
    });
    
    const data = await response.json();
    setUsers(data.items);
    setPagination({
      total: data.total,
      page: data.page,
      totalPages: data.total_pages
    });
  };

  return (
    <div>
      <h1>User Management</h1>
      
      {/* Filters */}
      <div className="filters">
        <input
          type="text"
          placeholder="Search..."
          value={filters.search}
          onChange={e => setFilters({...filters, search: e.target.value})}
        />
        <select
          value={filters.role}
          onChange={e => setFilters({...filters, role: e.target.value})}
        >
          <option value="">All Roles</option>
          <option value="user">User</option>
          <option value="admin">Admin</option>
          <option value="superadmin">Super Admin</option>
        </select>
        <select
          value={filters.is_active}
          onChange={e => setFilters({...filters, is_active: e.target.value})}
        >
          <option value="">All Status</option>
          <option value="true">Active</option>
          <option value="false">Inactive</option>
        </select>
      </div>

      {/* User Table */}
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Company</th>
            <th>Status</th>
            <th>Last Login</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {users.map(user => (
            <tr key={user.id}>
              <td>{user.first_name} {user.last_name}</td>
              <td>{user.email}</td>
              <td>{user.role}</td>
              <td>{user.company_name || '-'}</td>
              <td>{user.is_active ? 'Active' : 'Inactive'}</td>
              <td>{user.last_login_at || 'Never'}</td>
              <td>
                <button onClick={() => editUser(user.id)}>Edit</button>
                <button onClick={() => deleteUser(user.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Pagination */}
      <div className="pagination">
        Page {pagination.page} of {pagination.totalPages}
        <button
          disabled={filters.page === 1}
          onClick={() => setFilters({...filters, page: filters.page - 1})}
        >
          Previous
        </button>
        <button
          disabled={filters.page === pagination.totalPages}
          onClick={() => setFilters({...filters, page: filters.page + 1})}
        >
          Next
        </button>
      </div>
    </div>
  );
}
```

---

## Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

### Common Error Codes

- `unauthorized`: Missing or invalid authentication
- `forbidden`: Insufficient permissions
- `validation_error`: Invalid request data
- `not_found`: Resource not found
- `invalid_credentials`: Login failed
- `token_expired`: JWT token expired (trigger refresh)
- `csrf_token_invalid`: CSRF token mismatch
- `password_hashing_failed`: Server error during password hashing

### HTTP Status Codes

- `200 OK`: Successful GET/PUT request
- `201 Created`: Successful POST (creation)
- `204 No Content`: Successful DELETE or action with no response body
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource doesn't exist
- `500 Internal Server Error`: Server-side error

---

## Best Practices

1. **Always include credentials**: Use `credentials: 'include'` in fetch requests
2. **Store CSRF token securely**: Use localStorage or sessionStorage
3. **Handle 401 errors**: Implement automatic token refresh
4. **Validate permissions client-side**: Hide UI elements user can't access
5. **Use debouncing**: For search inputs to reduce API calls
6. **Cache user data**: Store current user info to avoid repeated /me calls
7. **Implement loading states**: Show spinners during API calls
8. **Handle errors gracefully**: Show user-friendly error messages
9. **Use optimistic updates**: Update UI before API response (with rollback)
10. **Implement pagination**: Don't load all data at once

---

## Support & Contact

For questions or issues, please contact the development team or open an issue in the repository.
