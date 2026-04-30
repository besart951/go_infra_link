# Frontend Architecture Notes

## Page Shape

- Route pages own route-level layout, headings, tab selection, and wiring.
- Complex route state belongs in small Svelte rune state classes or composables near the feature.
- Pure mapping, validation, filtering, payload building, and save reconciliation belongs in plain helper modules with unit tests.
- Presentational sections should receive state and callbacks as props. They should not fetch data unless they are a narrow lookup/select component.

## Repository Access

- Frontend application code should prefer domain repository ports and focused endpoint modules.
- Broad legacy adapters are migration scaffolding only. Do not add new production imports from broad facility adapters.
- Shared list endpoints should use the common list query builder and paginated response mapper where the backend contract matches `page`, `limit`, `search`, and simple filters.
- Route pages should import UI/state/use-case modules first. Direct infrastructure imports in pages are acceptable only for thin route-owned orchestration that has not yet moved into a state/use-case module.

## Translation Keys

- User-facing text should use `createTranslator()` and `$t('section.key')`.
- Add feature-specific keys under the closest existing namespace, for example `facility.alarm_catalog_page`.
- Keep table headings, empty states, button text, dialog text, and toast text translated.
- Existing hardcoded operational previews can be migrated incrementally, but new text should not add more hardcoded German strings.

## Static SPA Constraints

- Keep the current SvelteKit static adapter model.
- Do not introduce server hooks, server-only environment variables, or runtime-only server dependencies.
- Browser-only APIs must remain guarded by component lifecycle or client-side utilities.
- Repository modules should keep backend contracts unchanged during frontend refactors.
