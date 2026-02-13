# i18n - Internationalization (Localization) Guide

## Overview

This backend includes a clean, scalable internationalization (i18n) system for managing translations. Currently, only German (Switzerland) `de_CH` is needed, but the system is designed to easily support additional languages.

## Architecture

### Components

1. **Translator** (`pkg/i18n/translator.go`)
   - Core service for retrieving translated strings
   - Supports parameter substitution and pluralization
   - Automatic fallback to default language

2. **Loader** (`pkg/i18n/loader.go`)
   - Loads translation files from the filesystem
   - Parses JSON translation files
   - Supports loading all available languages

3. **Locale Middleware** (`internal/handler/middleware/locale.go`)
   - Automatically detects language from `Accept-Language` header
   - Stores locale and translator in request context
   - Converts HTTP language tags to internal format (e.g., `de-CH` → `de_CH`)

4. **Localized Error Handling** (`internal/handlerutil/i18n.go`)
   - Helpers for responding with localized error messages
   - Automatic translation of validation errors
   - Consistent error response format

## Translation Files

Translation files are stored in `internal/locales/` as JSON files with the naming convention:

- `{language_code}.json`

Examples:

- `de_CH.json` - German (Switzerland)
- `en_US.json` - English (USA)
- `fr_FR.json` - French (France)

### File Structure

Each translation file is organized by namespaces (nested JSON objects):

```json
{
  "common": {
    "success": "Erfolgreich",
    "error": "Ein Fehler ist aufgetreten"
  },
  "errors": {
    "not_found": "Die angeforderte Ressource wurde nicht gefunden.",
    "unauthorized": "Sie sind nicht autorisiert."
  },
  "auth": {
    "invalid_credentials": "E-Mail oder Passwort sind ungültig."
  }
}
```

Keys are flattened using dot notation:

- `common.success`
- `errors.not_found`
- `auth.invalid_credentials`

## Usage Examples

### In Handlers

#### Respond with Localized Error

```go
// Simple error with automatic translation
handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "not found")

// Or with a pre-translated message
message := translator.Get(locale, "errors.not_found")
handlerutil.RespondError(c, http.StatusNotFound, "not_found", message)
```

#### Localized Validation Errors

```go
// Automatic translation of validation errors
fields := handlerutil.LocalizeValidationErrors(c, &request, validationErrors)
handlerutil.RespondLocalizedValidationError(c, fields)
```

#### Single Field Error

```go
handlerutil.RespondLocalizedValidationFieldError(c, "email", "validation.email_invalid")
```

#### Custom Localized Response

```go
translator := middleware.GetTranslator(c)
locale := middleware.GetLocale(c)
message := translator.Get(locale, "auth.invalid_credentials")
handlerutil.RespondError(c, http.StatusUnauthorized, "invalid_credentials", message)
```

### Parameter Substitution

Translation strings can include placeholders:

In `de_CH.json`:

```json
{
  "validation": {
    "required": "{field} ist erforderlich.",
    "min_length": "{field} muss mindestens {min} Zeichen lang sein."
  }
}
```

In handler:

```go
message := translator.GetWithParams(locale, "validation.required", map[string]string{
  "field": "Email",
})
// Result: "Email ist erforderlich."

message := translator.GetWithParams(locale, "validation.min_length", map[string]string{
  "field": "Passwort",
  "min": "8",
})
// Result: "Passwort muss mindestens 8 Zeichen lang sein."
```

### Pluralization

For plural forms, add keys with `.zero`, `.one`, and `.other` suffixes:

In `de_CH.json`:

```json
{
  "items": {
    "count.zero": "Keine Einträge gefunden.",
    "count.one": "1 Eintrag gefunden.",
    "count.other": "{count} Einträge gefunden."
  }
}
```

In handler:

```go
message := translator.GetPlural(locale, "items.count", 0)   // "Keine Einträge gefunden."
message := translator.GetPlural(locale, "items.count", 1)   // "1 Eintrag gefunden."
message := translator.GetPlural(locale, "items.count", 42)  // "Einträge gefunden."
```

## Adding a New Language

To add support for a new language (e.g., English):

1. Create a new translation file: `internal/locales/en_US.json`
2. Add all keys from the default language, translated appropriately
3. The system will automatically load it on startup
4. Clients can request it via the `Accept-Language` header:
   ```
   Accept-Language: en-US
   ```

## How Language Detection Works

The system detects languages from the HTTP `Accept-Language` header:

```
Accept-Language: de-CH,de;q=0.9,en-US;q=0.8
```

Priority order:

1. `de-CH` (first, highest quality)
2. `de` (fallback)
3. `en-US` (fallback)

The middleware:

1. Parses the header (handles quality factors)
2. Normalizes the tag (`de-CH` → `de_CH`)
3. Stores it in the request context
4. Falls back to `de_CH` if no header or unmapped language

## API Error Response Format

All errors follow this consistent JSON format:

```json
{
  "error": "error_code",
  "message": "Localized error message"
}
```

For validation errors:

```json
{
  "error": "validation_error",
  "fields": {
    "email": "E-Mail muss eine gültige E-Mail-Adresse sein.",
    "password": "Passwort ist erforderlich."
  }
}
```

## Fallback Behavior

The translator implements intelligent fallback:

1. **Exact match**: `de_CH` language, specific key
2. **Default language**: Falls back to `de_CH` if different language requested
3. **Key as fallback**: Returns the key itself if no translation found

This ensures the API always returns meaningful responses.

## Configuration

The default language is set in `internal/app/app.go`:

```go
middleware.LocaleMiddleware(translator, "de_CH")
```

To change the default language, modify this line and ensure the corresponding translation file exists.

## Best Practices

1. **Use consistent key naming**: `feature.action.detail`
   - ✓ Good: `user.create.success`, `auth.login.invalid_credentials`
   - ✗ Avoid: `userCreateSuccess`, `authloginError`

2. **Parameterize messages**: Use `{placeholders}` for dynamic content
   - ✓ Good: `"Field {field} is required"`
   - ✗ Avoid: Hardcoded values in translations

3. **Group related translations**: Use namespaces to organize keys
   - ✓ Good: `auth.login`, `auth.logout`, `auth.password_reset`
   - ✗ Avoid: Flat structure with long key names

4. **Keep messages concise**: Translations should be readable in UI/responses
5. **Test with different Accept-Language headers**: Ensure fallback works

## Testing

To test with different languages in requests:

```bash
# Default (de_CH)
curl http://localhost:8080/api/v1/endpoint

# English (US)
curl -H "Accept-Language: en-US" http://localhost:8080/api/v1/endpoint

# French with fallback
curl -H "Accept-Language: fr-FR" http://localhost:8080/api/v1/endpoint
```

## Adding Translations to New Features

When adding a new feature:

1. Identify all user-facing messages and error codes
2. Add them to `internal/locales/de_CH.json` with appropriate keys
3. Use consistent namespacing (e.g., `feature_name.*` for all related messages)
4. Use helper functions in `handlerutil/i18n.go` to respond with localized messages
5. When adding a new language, translate the same keys

## Future Enhancements

- Database-backed translations for dynamic updates without redeployment
- Translation management UI
- Automated translation imports from services like Crowdin
- Caching optimizations for high-traffic scenarios
- Support for right-to-left languages
