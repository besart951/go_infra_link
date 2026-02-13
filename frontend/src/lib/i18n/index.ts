import { writable } from 'svelte/store';
import de_CH from './translations/de_CH.json' with { type: 'json' };

export type Locale = 'de-CH';

interface TranslationStore {
	locale: Locale;
	translations: Record<string, any>;
}

const translations: Record<Locale, Record<string, any>> = {
	'de-CH': de_CH
};

function createI18nStore() {
	const defaultLocale: Locale = 'de-CH';
	const { subscribe, set, update } = writable<TranslationStore>({
		locale: defaultLocale,
		translations: translations[defaultLocale]
	});

	return {
		subscribe,
		setLocale: (locale: Locale) => {
			update(store => ({
				...store,
				locale,
				translations: translations[locale] || translations[defaultLocale]
			}));
		},
		getTranslation: (key: string): string => {
			const parts = key.split('.');
			let current: any = translations[defaultLocale];
			
			for (const part of parts) {
				if (current && typeof current === 'object' && part in current) {
					current = current[part];
				} else {
					return key; // Return key if translation not found
				}
			}
			
			return typeof current === 'string' ? current : key;
		}
	};
}

export const i18n = createI18nStore();

/**
 * Helper function to get translation by key
 * Usage: t('auth.login') returns 'Anmelden'
 */
export function t(key: string): string {
	return i18n.getTranslation(key);
}
