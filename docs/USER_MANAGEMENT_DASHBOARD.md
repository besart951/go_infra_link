# User Management Dashboard

A high-fidelity user management dashboard with a clean, modern UI inspired by Clerk and Stripe.

## Features

### User Table
- Displays all users with the following columns:
  - **Name/Email**: User's full name and email address
  - **Role**: User role with visual badge (User, Admin, Super Admin)
  - **Status**: Account status (Active, Disabled, Locked)
  - **Last Active**: Relative time since last login (e.g., "2 days ago", "Yesterday")

### Search & Filtering
- Real-time search across user names and emails
- Instant feedback with results updating as you type

### Sorting
- Click on column headers to sort by:
  - Name (first_name)
  - Role
  - Last Active (last_login_at - default, descending)
- Visual indicators show current sort column and direction (ascending/descending)
- Default: Sorted by Last Login date (descending) - most recently active users appear first

### User Actions
Each user row has an actions menu (⋮) with the following options:

#### Change Role
- User
- Admin
- Super Admin

#### Account Management
- Enable User (if currently disabled)
- Disable User (if currently active)

#### Delete User
- Requires confirmation via dialog
- Irreversible action

### UI/UX Features
- **Loading States**: Skeleton loaders while data is fetching
- **Empty States**: Clean message when no users are found
- **Toast Notifications**: Success/error feedback for actions
- **Confirmation Dialogs**: For destructive actions like user deletion
- **Pagination**: Navigate through pages of users
- **Responsive Design**: Works on various screen sizes

## Technical Implementation

### Backend API
- **Endpoint**: `GET /api/v1/users`
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `limit`: Results per page (default: 10, max: 100)
  - `search`: Search term
  - `order_by`: Sort column (last_login_at, created_at, first_name, last_name, email, role)
  - `order`: Sort direction (asc, desc)

### Frontend Components
- `Table`: Custom table component built with shadcn-svelte patterns
- `Badge`: Color-coded badges for roles and statuses
- `Toast`: Toast notification system for user feedback
- `ConfirmDialog`: Confirmation dialog for destructive actions

### API Integration
All user management operations are handled through the API service layer:
- `listUsers()`: Fetch paginated user list
- `setUserRole()`: Change user role
- `disableUser()`: Disable a user account
- `enableUser()`: Enable a user account
- `deleteUser()`: Delete a user

## Security
- CodeQL security scan: **Passed (0 vulnerabilities)**
- SQL injection protection via column name whitelisting
- Proper authentication and authorization on backend routes
- Only admin users can access the user management dashboard

## Testing

To test the user management dashboard:

1. **Start the backend**:
   ```bash
   cd backend
   go run cmd/app/main.go
   ```

2. **Start the frontend**:
   ```bash
   cd frontend
   pnpm install
   pnpm run dev
   ```

3. **Access the dashboard**:
   - Navigate to `/users` after logging in as an admin user
   - Test search functionality by typing user names or emails
   - Test sorting by clicking column headers
   - Test user actions via the actions menu (⋮)

## Future Enhancements
- Team filtering (filter users by team membership)
- Bulk operations (select multiple users for bulk actions)
- Advanced filters (filter by role, status, date ranges)
- Export users to CSV
- User activity logs
