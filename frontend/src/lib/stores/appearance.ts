import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export type ThemePreference = 'system' | 'light' | 'dark';
export type FontPreference = 'noto' | 'system' | 'serif' | 'mono';

export const MIN_CONTRAST_PREFERENCE = 80;
export const MAX_CONTRAST_PREFERENCE = 140;
export const CONTRAST_PREFERENCE_STEP = 5;
export const DEFAULT_CONTRAST_PREFERENCE = 100;
export const DEFAULT_FONT_PREFERENCE: FontPreference = 'noto';

export const FONT_STACKS: Record<FontPreference, string> = {
  noto: "'Noto Sans Variable', 'Noto Sans', sans-serif",
  system: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
  serif: "Georgia, 'Times New Roman', serif",
  mono: "'Cascadia Code', 'SFMono-Regular', Consolas, 'Liberation Mono', monospace"
};

const THEME_STORAGE_KEY = 'theme_preference';
const CONTRAST_STORAGE_KEY = 'contrast_preference';
const FONT_STORAGE_KEY = 'font_preference';

function isThemePreference(value: unknown): value is ThemePreference {
  return value === 'system' || value === 'light' || value === 'dark';
}

function isFontPreference(value: unknown): value is FontPreference {
  return value === 'noto' || value === 'system' || value === 'serif' || value === 'mono';
}

function normalizeContrastPreference(value: unknown): number {
  const numeric = typeof value === 'number' ? value : Number(value);
  if (!Number.isFinite(numeric)) return DEFAULT_CONTRAST_PREFERENCE;

  const clamped = Math.min(MAX_CONTRAST_PREFERENCE, Math.max(MIN_CONTRAST_PREFERENCE, numeric));

  return Math.round(clamped / CONTRAST_PREFERENCE_STEP) * CONTRAST_PREFERENCE_STEP;
}

function readStoredThemePreference(): ThemePreference {
  if (!browser) return 'system';
  try {
    const raw = localStorage.getItem(THEME_STORAGE_KEY);
    return isThemePreference(raw) ? raw : 'system';
  } catch {
    return 'system';
  }
}

function readStoredContrastPreference(): number {
  if (!browser) return DEFAULT_CONTRAST_PREFERENCE;
  try {
    return normalizeContrastPreference(localStorage.getItem(CONTRAST_STORAGE_KEY));
  } catch {
    return DEFAULT_CONTRAST_PREFERENCE;
  }
}

function readStoredFontPreference(): FontPreference {
  if (!browser) return DEFAULT_FONT_PREFERENCE;
  try {
    const raw = localStorage.getItem(FONT_STORAGE_KEY);
    return isFontPreference(raw) ? raw : DEFAULT_FONT_PREFERENCE;
  } catch {
    return DEFAULT_FONT_PREFERENCE;
  }
}

function prefersDark(): boolean {
  return window.matchMedia?.('(prefers-color-scheme: dark)')?.matches ?? false;
}

function applyThemePreferenceToDom(preference: ThemePreference) {
  const dark = preference === 'dark' || (preference === 'system' && prefersDark());
  document.documentElement.classList.toggle('dark', dark);
}

function applyContrastPreferenceToDom(preference: number) {
  document.documentElement.style.setProperty('--app-contrast', `${preference}%`);
  document.documentElement.style.setProperty(
    '--app-contrast-filter',
    preference === DEFAULT_CONTRAST_PREFERENCE ? 'none' : `contrast(${preference}%)`
  );
}

function applyFontPreferenceToDom(preference: FontPreference) {
  document.documentElement.style.setProperty('--app-font-family', FONT_STACKS[preference]);
}

export const themePreference = writable<ThemePreference>('system');
export const contrastPreference = writable<number>(DEFAULT_CONTRAST_PREFERENCE);
export const fontPreference = writable<FontPreference>(DEFAULT_FONT_PREFERENCE);

let initialized = false;
let mediaQuery: MediaQueryList | null = null;
let currentThemePreference: ThemePreference = 'system';

export function setThemePreference(preference: ThemePreference) {
  themePreference.set(preference);
}

export function setContrastPreference(preference: number) {
  contrastPreference.set(normalizeContrastPreference(preference));
}

export function setFontPreference(preference: FontPreference) {
  fontPreference.set(preference);
}

export function initAppearance() {
  if (!browser || initialized) return;
  initialized = true;

  const storedThemePreference = readStoredThemePreference();
  const storedContrastPreference = readStoredContrastPreference();
  const storedFontPreference = readStoredFontPreference();

  currentThemePreference = storedThemePreference;
  themePreference.set(storedThemePreference);
  contrastPreference.set(storedContrastPreference);
  fontPreference.set(storedFontPreference);

  applyThemePreferenceToDom(storedThemePreference);
  applyContrastPreferenceToDom(storedContrastPreference);
  applyFontPreferenceToDom(storedFontPreference);

  mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)') ?? null;
  mediaQuery?.addEventListener('change', () => {
    if (currentThemePreference === 'system') applyThemePreferenceToDom('system');
  });

  themePreference.subscribe((pref) => {
    currentThemePreference = pref;
    applyThemePreferenceToDom(pref);
    try {
      localStorage.setItem(THEME_STORAGE_KEY, pref);
    } catch {
      // ignore
    }
  });

  contrastPreference.subscribe((pref) => {
    const normalized = normalizeContrastPreference(pref);
    applyContrastPreferenceToDom(normalized);
    try {
      localStorage.setItem(CONTRAST_STORAGE_KEY, String(normalized));
    } catch {
      // ignore
    }
  });

  fontPreference.subscribe((pref) => {
    applyFontPreferenceToDom(pref);
    try {
      localStorage.setItem(FONT_STORAGE_KEY, pref);
    } catch {
      // ignore
    }
  });
}

export const initTheme = initAppearance;
