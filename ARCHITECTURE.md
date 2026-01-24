# Architecture Diagram

## Request Flow: Frontend → Backend

```
┌─────────────────────────────────────────────────────────────────┐
│                     Browser / SvelteKit App                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. User clicks "Create Team"                                   │
│     ↓                                                             │
│  2. Component calls: await createTeam({ name: 'My Team' })      │
│     (from $lib/api/teams.ts)                                     │
│     ↓                                                             │
│  3. Function calls: api('/teams', { method: 'POST', ... })      │
│     (from $lib/api/client.ts)                                    │
│     ↓                                                             │
│  4. api() function:                                              │
│     • Reads csrf_token from cookies ✓                           │
│     • Adds X-CSRF-Token header ✓                                │
│     • Adds credentials: 'include' ✓                             │
│     ↓                                                             │
│  5. HTTP POST /api/v1/teams                                      │
│     Headers:                                                      │
│     - X-CSRF-Token: aa7b4fdc...                                 │
│     - Cookie: access_token=...; csrf_token=...                  │
│     - Content-Type: application/json                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                             ↓ Network
┌─────────────────────────────────────────────────────────────────┐
│                        Go Backend                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Router receives POST /api/v1/teams                          │
│     ↓                                                             │
│  2. CSRF middleware validates X-CSRF-Token header              │
│     ↓                                                             │
│  3. Auth middleware checks access_token cookie                 │
│     ↓                                                             │
│  4. Handler processes request                                   │
│     ↓                                                             │
│  5. Returns 200 OK + JSON response                              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                             ↓ Network
┌─────────────────────────────────────────────────────────────────┐
│                      Browser again                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  6. Response received & parsed                                  │
│     ↓                                                             │
│  7. Component updates UI                                        │
│     ↓                                                             │
│  8. Done! ✓                                                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration Flow

```
┌──────────────────┐
│   Root .env      │  Shared configuration
│  (repo root)     │  for backend, frontend, docker
└────────┬─────────┘
         │
    ┌────┴────────────────────────────────┐
    │                                     │
    ↓                                     ↓
┌────────────────┐            ┌────────────────────┐
│ Backend        │            │ Frontend (SvelteKit)│
│ (Go)           │            │                    │
├────────────────┤            ├────────────────────┤
│ • config.Load()│            │ • $env/dynamic/*   │
│   reads:       │            │ • getBackendUrl()  │
│   - APP_ENV    │            │ • api() client     │
│   - JWT_SECRET │            │                    │
│   - DB_DRIVER  │            │ Uses:              │
│   - etc.       │            │ - GDK_BACKEND_URL  │
└────────────────┘            │ - GDK_BACKEND_HOST │
    │                         │ - APP_ENV          │
    │                         └────────────────────┘
    │
    ↓
┌─────────────────────┐
│ Docker Compose      │
│ env_file: .env      │
└─────────────────────┘
```

## CSRF Token Flow

```
┌──────────────────────────────────────────────┐
│           1. User Logs In                     │
├──────────────────────────────────────────────┤
│ POST /api/v1/auth/login                      │
│ { email, password }                          │
└──────────────┬───────────────────────────────┘
               │
               ↓ Backend validates credentials
┌──────────────────────────────────────────────┐
│         2. Backend Response                   │
├──────────────────────────────────────────────┤
│ 200 OK                                       │
│ Set-Cookie: access_token=...                │
│ Set-Cookie: refresh_token=...               │
│ Set-Cookie: csrf_token=aa7b4fdc... ✓        │
│ Body: { csrf_token, access_token_expires... }│
└──────────────┬───────────────────────────────┘
               │
               ↓ Browser stores cookies
┌──────────────────────────────────────────────┐
│      3. Cookies Stored in Browser            │
├──────────────────────────────────────────────┤
│ access_token (httpOnly)                      │
│ refresh_token (httpOnly)                     │
│ csrf_token (not httpOnly - js can read) ✓   │
└──────────────┬───────────────────────────────┘
               │
               ↓ api() client uses cookie
┌──────────────────────────────────────────────┐
│    4. Make Protected Request                  │
├──────────────────────────────────────────────┤
│ const team = await api('/teams', {           │
│   method: 'POST',                            │
│   body: JSON.stringify({ name })             │
│ })                                           │
│                                              │
│ api() function:                              │
│ • Reads csrf_token from cookie ✓            │
│ • Adds X-CSRF-Token header ✓                │
│ • Includes credentials: 'include' ✓         │
└──────────────┬───────────────────────────────┘
               │
               ↓
┌──────────────────────────────────────────────┐
│   5. Backend Validates                        │
├──────────────────────────────────────────────┤
│ CSRF middleware checks:                      │
│ • Header X-CSRF-Token == cookie csrf_token  │
│ If match: Request proceeds ✓                │
│ If no match: 403 Forbidden ✗                │
└──────────────────────────────────────────────┘
```

## File Structure

```
go_infra_link/
├── .env                          ← Shared config (repo root)
├── SETUP.md                      ← Setup guide
├── docker-compose.yml            ← Uses .env
├── backend/
│   ├── internal/config/
│   │   └── config.go            ← Loads from ../.env
│   ├── cmd/app/
│   │   └── main.go
│   └── .env                     ← (optional) local override
│
└── frontend/
    ├── QUICK_REFERENCE.md       ← API usage examples
    ├── src/lib/
    │   ├── api/
    │   │   ├── client.ts        ← Core api() function + error handling
    │   │   ├── users.ts         ← High-level user functions
    │   │   ├── teams.ts         ← High-level team functions
    │   │   └── README.md        ← API docs
    │   │
    │   └── server/
    │       ├── backend.ts       ← getBackendUrl() helper
    │       └── set-cookie.ts    ← getSetCookieValues() parser
    │
    └── src/routes/
        ├── (auth)/login/
        │   ├── +page.svelte
        │   └── +page.server.ts  ← Uses getBackendUrl()
        │
        └── (app)/
            ├── +layout.server.ts ← Checks backend health
            └── ...
```

## Environment Variable Precedence

```
Backend (Go):
1. Environment variables (OS)
2. .env (backend directory) ← overrides below
3. ../.env (repo root) ← shared config
4. Default fallback values

Frontend (SvelteKit):
1. Environment variables (OS)
2. .env.local (vite)
3. $env/dynamic/private ← reads from .env files
4. Hard-coded fallbacks (getBackendUrl defaults)

Docker Compose:
1. env_file: .env ← shared config
2. environment: section ← inline overrides
```

## Error Handling Chain

```
Component
   ↓
try { await api('/teams', ...) } catch (err)
   │
   ├─→ ApiException (status 400-500)
   │   ├─ 400: Missing fields
   │   ├─ 401: Unauthorized
   │   ├─ 403: Forbidden
   │   ├─ 404: Not found
   │   └─ 500: Server error
   │
   ├─→ TypeError (network error)
   │   └─ "Network request failed"
   │
   └─→ Unknown error
       └─ "An unexpected error occurred"

Helper functions:
• getErrorMessage(err) → display-friendly string
• isApiError(err, 'unauthorized') → type checking
```
