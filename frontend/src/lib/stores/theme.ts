import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export type ThemePreference = 'system' | 'light' | 'dark';

const STORAGE_KEY = 'theme_preference';

function isThemePreference(value: unknown): value is ThemePreference {
	return value === 'system' || value === 'light' || value === 'dark';
}

function readStoredPreference(): ThemePreference {
	if (!browser) return 'system';
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return isThemePreference(raw) ? raw : 'system';
	} catch {
		return 'system';
	}
}

function prefersDark(): boolean {
	return window.matchMedia?.('(prefers-color-scheme: dark)')?.matches ?? false;
}

function applyPreferenceToDom(preference: ThemePreference) {
	const dark = preference === 'dark' || (preference === 'system' && prefersDark());
	document.documentElement.classList.toggle('dark', dark);
}

export const themePreference = writable<ThemePreference>('system');

let initialized = false;
let mediaQuery: MediaQueryList | null = null;
let currentPreference: ThemePreference = 'system';

export function setThemePreference(preference: ThemePreference) {
	themePreference.set(preference);
}

export function initTheme() {
	if (!browser || initialized) return;
	initialized = true;

	const preference = readStoredPreference();
		currentPreference = preference;
		themePreference.set(preference);
		applyPreferenceToDom(preference);

	mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)') ?? null;
		mediaQuery?.addEventListener('change', () => {
			if (currentPreference === 'system') applyPreferenceToDom('system');
		});

		themePreference.subscribe((pref) => {
			currentPreference = pref;
		applyPreferenceToDom(pref);
		try {
			localStorage.setItem(STORAGE_KEY, pref);
		} catch {
			// ignore
		}
	});
}
