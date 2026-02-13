# Translations

This directory contains all application translations that are served to the frontend via API.

## Architecture

The translations follow a professional client-server architecture:

- **Backend**: Serves translations via REST API (`/api/v1/i18n/:locale`)
- **Frontend**: Loads translations dynamically at runtime from the backend

This approach is used by professional applications because it provides:

- **Single source of truth**: All translations managed in one location (backend)
- **Runtime updates**: Translations can be updated without rebuilding the frontend
- **Centralized control**: Backend controls which translations are available
- **Consistency**: Frontend and backend always use the same translations
- **Easier deployment**: No need to sync files between frontend and backend

## File Structure

- `de_CH.json` - Swiss German translations (default language)

## API Endpoint

### GET `/api/v1/i18n/:locale`

Returns all translations for the specified locale.

**Example Request:**

```http
GET /api/v1/i18n/de_CH
```

**Example Response:**

```json
{
  "common": {
    "save": "Speichern",
    "cancel": "Abbrechen",
    ...
  },
  "auth": {
    "login": "Anmelden",
    ...
  }
}
```

## Frontend Integration

The frontend loads translations automatically on application start:

```typescript
// Frontend automatically fetches translations from backend
import { i18n } from "$lib/i18n/index.js";

// Shows loading state while translations are being fetched
// Shows error state if fetch fails with retry button
```

## Adding New Translations

1. Edit the `.json` file in this directory
2. Restart the backend server
3. Frontend will automatically load the new translations

## Translation Key Naming Convention

Use dot notation for hierarchical organization:

- `common.*` - Common UI elements (buttons, labels)
- `auth.*` - Authentication/authorization messages
- `errors.*` - Error messages
- `validation.*` - Form validation messages
- `user.*` - User management
- `team.*` - Team management
- `project.*` - Project management
- `facility.*` - Facility management
- `phase.*` - Phase management
- `navigation.*` - Navigation labels
- `pages.*` - Page-specific content
- `messages.*` - Toast/notification messages

## Locale Code: de_CH

This application uses `de_CH` (Swiss German / Schweizerdeutsch) as the primary locale.

- Use Swiss spelling conventions (e.g., "Schliessen" not "Schließen")
- Date formats follow Swiss conventions
- Currency and number formats use Swiss standards
