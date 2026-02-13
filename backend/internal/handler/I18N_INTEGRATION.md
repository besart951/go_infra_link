# i18n Integration Guide for Handlers

This guide shows practical examples of how to integrate localized messages into your handlers.

## Example: Auth Handler with Localization

```go
package handler

import (
    "github.com/besart951/go_infra_link/backend/internal/handler/middleware"
    "github.com/besart951/go_infra_link/backend/internal/handlerutil"
    "github.com/gin-gonic/gin"
)

// Login handler with localized error messages
func (h *AuthHandler) Login(c *gin.Context) {
    var req struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required,min=6"`
    }

    // Localized validation error handling
    if !handlerutil.BindJSON(c, &req) {
        return
    }

    user, err := h.service.Authenticate(req.Email, req.Password)

    if err != nil {
        // Respond with localized error message
        // The locale is automatically extracted from Accept-Language header
        handlerutil.RespondLocalizedError(
            c,
            http.StatusUnauthorized,
            "invalid_credentials",
            "auth.invalid_credentials",
        )
        return
    }

    // ... rest of login logic
    c.JSON(http.StatusOK, gin.H{"success": true})
}

// User creation with validation errors
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest

    // This will automatically translate validation errors
    if !handlerutil.BindJSON(c, &req) {
        return
    }

    user, err := h.service.Create(&req)
    if err != nil {
        if errors.Is(err, domain.ErrConflict) {
            handlerutil.RespondLocalizedError(
                c,
                http.StatusConflict,
                "conflict",
                "errors.conflict",
            )
            return
        }

        handlerutil.RespondLocalizedError(
            c,
            http.StatusInternalServerError,
            "internal_error",
            "errors.internal_server_error",
        )
        return
    }

    c.JSON(http.StatusCreated, user)
}
```

## Step-by-Step Integration

### 1. Add Error Messages to Translations

First, add your error messages to `internal/locales/de_CH.json`:

```json
{
  "feature_name": {
    "operation_success": "Operation war erfolgreich",
    "operation_failed": "Operation ist fehlgeschlagen"
  }
}
```

### 2. Import Required Packages

```go
import (
    "github.com/besart951/go_infra_link/backend/internal/handler/middleware"
    "github.com/besart951/go_infra_link/backend/internal/handlerutil"
)
```

### 3. Use Localization in Handlers

#### For Errors with Auto-Translation

```go
// Responds with localized message from "errors.not_found"
handlerutil.RespondLocalizedError(
    c,
    http.StatusNotFound,
    "not_found",
    "not found", // fallback if no translation
)

// Or more simply - the key is the same as the translation key
handlerutil.RespondLocalizedNotFound(c)
```

#### For Validation Errors

```go
if err := c.ShouldBindJSON(&req); err != nil {
    // This automatically translates validation messages
    fields := handlerutil.LocalizeValidationErrors(c, &req, validationErrors)
    handlerutil.RespondLocalizedValidationError(c, fields)
    return
}
```

#### For Custom Messages with Parameters

```go
translator := middleware.GetTranslator(c)
locale := middleware.GetLocale(c)

message := translator.GetWithParams(
    locale,
    "validation.required",
    map[string]string{"field": "Email"},
)
// Result: "Email ist erforderlich."

handlerutil.RespondError(c, http.StatusBadRequest, "validation", message)
```

#### For Dynamically Retrieved Messages

```go
translator := middleware.GetTranslator(c)
locale := middleware.GetLocale(c)

message := translator.Get(locale, "auth.invalid_credentials")
handlerutil.RespondError(c, http.StatusUnauthorized, "invalid_credentials", message)
```

## Common Patterns

### Pattern 1: Simple Error Response

```go
handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", "Entity not found")
```

### Pattern 2: Validation with Field Error

```go
fields := make(map[string]string)
fields["email"] = translator.Get(locale, "validation.email_invalid")
handlerutil.RespondLocalizedValidationError(c, fields)
```

### Pattern 3: Domain Error Mapping

```go
if errors.Is(err, domain.ErrNotFound) {
    handlerutil.RespondLocalizedNotFound(c)
    return
}

if errors.Is(err, domain.ErrConflict) {
    handlerutil.RespondLocalizedError(
        c,
        http.StatusConflict,
        "conflict",
        "errors.conflict",
    )
    return
}
```

### Pattern 4: Custom Message with Dynamic Values

```go
translator := middleware.GetTranslator(c)
locale := middleware.GetLocale(c)

params := map[string]string{
    "resource":  "User",
    "operation": "deleted",
}

message := translator.GetWithParams(locale, "messages.resource_action", params)
// Assuming translation: "{resource} wurde {operation}"
// Result: "User wurde deleted"

c.JSON(http.StatusOK, gin.H{"message": message})
```

## Testing Localized Handlers

### Test with Default Language (de_CH)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "123456"}'

# Response (with German error message):
# {
#   "error": "invalid_credentials",
#   "message": "E-Mail oder Passwort sind ungültig."
# }
```

### Test with Specific Language

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "Accept-Language: en-US" \
  -d '{"email": "test@example.com", "password": "123456"}'

# Will fall back to German (de_CH) if English not available
```

### Test with Validation Errors

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "invalid", "password": "123"}'

# Response (with German field error messages):
# {
#   "error": "validation_error",
#   "fields": {
#     "email": "Email muss eine gültige E-Mail-Adresse sein.",
#     "password": "Passwort muss mindestens 8 Zeichen lang sein."
#   }
# }
```

## Adding New Error Messages

1. Add to translation files in `internal/locales/de_CH.json`:

```json
{
  "your_feature": {
    "error_key": "Error message in German"
  }
}
```

2. Use in handler:

```go
handlerutil.RespondLocalizedError(
    c,
    http.StatusBadRequest,
    "error_key",
    "errors.your_feature.error_key",
)
```

## Accessing Locale Info

In your handlers, you can access locale information:

```go
// Get current locale (e.g., "de_CH")
locale := middleware.GetLocale(c)

// Get translator instance
translator := middleware.GetTranslator(c)

// Manually translate
message := translator.Get(locale, "any.translation.key")
```

## API Response Structure

All API responses follow this structure for errors:

```json
{
  "error": "error_code",
  "message": "Localized message based on Accept-Language header",
  "fields": {
    "field_name": "Field-specific error message"
  }
}
```

The `fields` key is only present for validation errors.

## Notes

- Locale detection is automatic via the `Accept-Language` request header
- Default fallback is `de_CH` (German Switzerland)
- If a translation is not found, the translation key is returned as-is
- Parameter substitution works with `{paramName}` placeholders in translation strings
- Validation error translation is automatic with `LocalizeValidationErrors()`
