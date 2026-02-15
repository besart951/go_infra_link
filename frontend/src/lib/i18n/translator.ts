import { i18n } from './index.js';
import type { Readable } from 'svelte/store';

export type TranslationParams = Record<string, string | number | boolean | null | undefined>;

/**
 * Reactive translation getter
 * Returns a Svelte store that provides the translation function
 * Usage in components:
 * <script>
 *   import { createTranslator } from '$lib/i18n/translator'
 *   const t = createTranslator()
 * </script>
 * <div>{$t('auth.login')}</div>
 */
export function createTranslator(): Readable<
	(key: string, params?: TranslationParams) => string
> {
	return {
		subscribe(fn) {
			return i18n.subscribe(() => {
				fn((key: string, params?: TranslationParams) => i18n.getTranslation(key, params));
			});
		}
	};
}
