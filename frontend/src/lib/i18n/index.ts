import { writable } from 'svelte/store';

export type Locale = 'de-CH';

interface TranslationStore {
	locale: Locale;
	translations: Record<string, any>;
	isLoading: boolean;
	error: string | null;
}

const translations: Record<Locale, Record<string, any>> = {
	'de-CH': {}
};

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function createI18nStore() {
	const defaultLocale: Locale = 'de-CH';
	let currentTranslations: Record<string, any> = {};

	const { subscribe, set, update } = writable<TranslationStore>({
		locale: defaultLocale,
		translations: {},
		isLoading: true,
		error: null
	});

	// Load translations from backend
	async function loadTranslations(locale: Locale): Promise<void> {
		update((store) => ({ ...store, isLoading: true, error: null }));

		try {
			// Convert locale format: 'de-CH' -> 'de_CH'
			const localeParam = locale.replace('-', '_');
			const response = await fetch(`${API_BASE_URL}/api/v1/i18n/${localeParam}`);

			if (!response.ok) {
				throw new Error(`Failed to load translations: ${response.statusText}`);
			}

			const data = await response.json();
			translations[locale] = data;
			currentTranslations = data;

			update((store) => ({
				...store,
				locale,
				translations: data,
				isLoading: false,
				error: null
			}));
		} catch (err) {
			const errorMsg = err instanceof Error ? err.message : 'Failed to load translations';
			update((store) => ({
				...store,
				isLoading: false,
				error: errorMsg
			}));
			console.error('Failed to load translations:', err);
		}
	}

	// Initialize by loading default locale
	loadTranslations(defaultLocale);

	return {
		subscribe,
		setLocale: async (locale: Locale) => {
			if (!translations[locale] || Object.keys(translations[locale]).length === 0) {
				await loadTranslations(locale);
			} else {
				currentTranslations = translations[locale];
				update((store) => ({
					...store,
					locale,
					translations: translations[locale]
				}));
			}
		},
		reload: () => {
			let currentLocale: Locale = defaultLocale;
			update((store) => {
				currentLocale = store.locale;
				return store;
			});
			return loadTranslations(currentLocale);
		},
		getTranslation: (key: string): string => {
			const parts = key.split('.');
			let current: any = currentTranslations;

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
