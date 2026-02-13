# Frontend i18n (Internationalization) Documentation

## Overview

The frontend uses a Svelte store-based translation system for German (Switzerland) localization. All user-facing text should use translation keys instead of hardcoded strings.

## Location

- **Translation files**: `src/lib/i18n/translations/de_CH.json`
- **i18n store**: `src/lib/i18n/index.ts`
- **Translator utility**: `src/lib/i18n/translator.ts`

## Usage in Svelte Components

### Method 1: Using createTranslator Hook (Recommended)

```svelte
<script lang="ts">
  import { createTranslator } from '$lib/i18n/translator';
  
  const t = createTranslator();
</script>

<h1>{$t('common.app_name')}</h1>
<button>{$t('auth.login')}</button>
<p>{$t('facility.building_not_found')}</p>
```

### Method 2: Direct Import

```svelte
<script lang="ts">
  import { t } from '$lib/i18n';
</script>

<h1>{t('common.app_name')}</h1>
<!-- Note: This approach is NOT reactive to locale changes -->
```

## Translation Key Structure

Translation keys are organized hierarchically in `de_CH.json`:

```json
{
  "section": {
    "key": "German translation"
  }
}
```

### Available Sections

- `common` - Common UI elements (buttons, navigation)
- `auth` - Authentication messages (login, signup, errors)
- `user` - User management
- `team` - Team management
- `project` - Project management
- `facility` - Facility/infrastructure management
- `phase` - Project phases
- `errors` - Generic error messages
- `navigation` - Navigation menu
- `messages` - Notification messages

### Usage Examples

```
auth.login              → "Anmelden"
auth.invalid_credentials → "Ungültige E-Mail oder Passwort."
facility.building       → "Gebäude"
facility.building_not_found → "Gebäude nicht gefunden."
common.save            → "Speichern"
errors.server_error    → "Serverfehler. Bitte versuchen Sie es später erneut."
```

## Adding New Translations

1. Open `src/lib/i18n/translations/de_CH.json`
2. Add your new key-value pair under the appropriate section
3. Use the key in your component with `$t('section.key')`

Example:
```json
{
  "facility": {
    "my_new_message": "Meine neue Nachricht"
  }
}
```

Then in a component:
```svelte
<p>{$t('facility.my_new_message')}</p>
```

## Current Translation Coverage

✅ **Completed:**
- Authentication (login, signup, password reset)
- Common UI elements
- Error messages
- Basic facility management
- User & team management
- Project management

⏳ **To Be Translated:**
- Component-specific messages
- Dialog/modal content
- Validation messages
- Toast notifications

## Best Practices

1. **Never use hardcoded strings** - Always use translation keys
2. **Be consistent** - Use existing keys when available
3. **Group related keys** - Keep similar messages in the same section
4. **Use descriptive keys** - e.g., `building_not_found` is better than `error_1`
5. **Lazy translation** - Only translate strings that are user-facing

## Fallback Behavior

If a translation key is not found, the key itself is returned:
- Missing key: `$t('unknown.key')` → `'unknown.key'`
- This makes it easy to spot untranslated content during development

## Future Enhancements

When adding new locale support (e.g., English):
1. Create `src/lib/i18n/translations/en_US.json`
2. Update the `translations` object in `src/lib/i18n/index.ts`
3. Add locale option to `Locale` type
4. Update `setLocale()` function to accept new locale
