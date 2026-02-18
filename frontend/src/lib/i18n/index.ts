import { writable } from 'svelte/store';
import deCHFallback from './translations/de_CH.json';

export type Locale = 'de-CH';
export type TranslationParams = Record<string, string | number | boolean | null | undefined>;

interface TranslationStore {
	locale: Locale;
	translations: Record<string, any>;
	isLoading: boolean;
	error: string | null;
}

const translations: Record<Locale, Record<string, any>> = {
	'de-CH': {}
};

const localTranslations: Record<Locale, Record<string, any>> = {
	'de-CH': deCHFallback as Record<string, any>
};

const API_BASE_PATH = '/api/v1';

function createI18nStore() {
	const defaultLocale: Locale = 'de-CH';
	let currentTranslations: Record<string, any> = {};

	function deepMerge(
		base: Record<string, any>,
		override: Record<string, any>
	): Record<string, any> {
		const result: Record<string, any> = { ...base };
		for (const [key, value] of Object.entries(override)) {
			if (
				value &&
				typeof value === 'object' &&
				!Array.isArray(value) &&
				base[key] &&
				typeof base[key] === 'object' &&
				!Array.isArray(base[key])
			) {
				result[key] = deepMerge(base[key], value as Record<string, any>);
			} else {
				result[key] = value;
			}
		}
		return result;
	}

	const { subscribe, set, update } = writable<TranslationStore>({
		locale: defaultLocale,
		translations: {},
		isLoading: true,
		error: null
	});

	// Load translations from backend
	async function loadTranslations(locale: Locale): Promise<void> {
		update((store) => ({ ...store, isLoading: true, error: null }));

		const fallbackData = localTranslations[locale];
		const hasFallback = Boolean(fallbackData && Object.keys(fallbackData).length > 0);

		if (hasFallback) {
			translations[locale] = fallbackData;
			currentTranslations = fallbackData;
			update((store) => ({
				...store,
				locale,
				translations: fallbackData,
				isLoading: false,
				error: null
			}));
		}

		try {
			// Convert locale format: 'de-CH' -> 'de_CH'
			const localeParam = locale.replace('-', '_');
			const response = await fetch(`${API_BASE_PATH}/i18n/${localeParam}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`Failed to load translations: ${response.statusText}`);
			}

			const data = await response.json();
			const merged = hasFallback ? deepMerge(fallbackData, data) : data;
			translations[locale] = merged;
			currentTranslations = merged;

			update((store) => ({
				...store,
				locale,
				translations: merged,
				isLoading: false,
				error: null
			}));
		} catch (err) {
			if (hasFallback) {
				console.warn('Failed to refresh translations from backend, using local fallback:', err);
				return;
			}

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
		getTranslation: (key: string, params?: TranslationParams): string => {
			return getTranslationWithParams(key, params);
		}
	};

	function interpolate(template: string, params?: TranslationParams): string {
		if (!params) return template;

		let result = template;
		for (const [name, rawValue] of Object.entries(params)) {
			const value = String(rawValue ?? '');
			result = result
				.replaceAll(`{${name}}`, value)
				.replaceAll(`{{${name}}}`, value)
				.replaceAll(`:${name}`, value);
		}

		return result;
	}

	function getTranslationWithParams(key: string, params?: TranslationParams): string {
		const parts = key.split('.');
		let current: any = currentTranslations;

		for (const part of parts) {
			if (current && typeof current === 'object' && part in current) {
				current = current[part];
			} else {
				return key; // Return key if translation not found
			}
		}

		return typeof current === 'string' ? interpolate(current, params) : key;
	}
}

export const i18n = createI18nStore();

/**
 * Helper function to get translation by key
 * Usage: t('auth.login') returns 'Anmelden'
 */
export function t(key: string, params?: TranslationParams): string {
	return i18n.getTranslation(key, params);
}
