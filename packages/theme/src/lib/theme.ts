import { browser } from "$app/environment";
import { writable } from "svelte/store";

export type ThemePreference = "system" | "light" | "dark";
export type ContrastPreference = 100 | 110 | 120 | 130;

const STORAGE_KEY = "theme_preference";
const CONTRAST_STORAGE_KEY = "contrast_preference";
const DEFAULT_CONTRAST: ContrastPreference = 100;

function isThemePreference(value: unknown): value is ThemePreference {
  return value === "system" || value === "light" || value === "dark";
}

function isContrastPreference(value: unknown): value is ContrastPreference {
  return value === 100 || value === 110 || value === 120 || value === 130;
}

function contrastStorageKey(userId: string): string {
  return `${CONTRAST_STORAGE_KEY}:${userId}`;
}

function readStoredPreference(): ThemePreference {
  if (!browser) return "system";
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return isThemePreference(raw) ? raw : "system";
  } catch {
    return "system";
  }
}

function readStoredContrast(userId: string): ContrastPreference {
  if (!browser) return DEFAULT_CONTRAST;
  try {
    const raw = localStorage.getItem(contrastStorageKey(userId));
    const value = raw === null ? DEFAULT_CONTRAST : Number(raw);
    return isContrastPreference(value) ? value : DEFAULT_CONTRAST;
  } catch {
    return DEFAULT_CONTRAST;
  }
}

function prefersDark(): boolean {
  return window.matchMedia?.("(prefers-color-scheme: dark)")?.matches ?? false;
}

function applyPreferenceToDom(preference: ThemePreference) {
  const dark =
    preference === "dark" || (preference === "system" && prefersDark());
  document.documentElement.classList.toggle("dark", dark);
}

function applyContrastToDom(preference: ContrastPreference) {
  document.documentElement.style.setProperty("--app-contrast", `${preference}%`);
}

export const themePreference = writable<ThemePreference>("system");
export const contrastPreference = writable<ContrastPreference>(DEFAULT_CONTRAST);

let initialized = false;
let mediaQuery: MediaQueryList | null = null;
let currentPreference: ThemePreference = "system";
let currentUserId: string | null = null;

export function setThemePreference(preference: ThemePreference) {
  themePreference.set(preference);
}

export function setContrastPreference(preference: ContrastPreference) {
  contrastPreference.set(preference);
}

export function initTheme() {
  if (!browser || initialized) return;
  initialized = true;

  const preference = readStoredPreference();
  currentPreference = preference;
  themePreference.set(preference);
  applyPreferenceToDom(preference);

  mediaQuery = window.matchMedia?.("(prefers-color-scheme: dark)") ?? null;
  mediaQuery?.addEventListener("change", () => {
    if (currentPreference === "system") applyPreferenceToDom("system");
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

  contrastPreference.subscribe((pref) => {
    applyContrastToDom(pref);
    if (!currentUserId) return;

    try {
      localStorage.setItem(contrastStorageKey(currentUserId), String(pref));
    } catch {
      // ignore
    }
  });
}

export function initUserPreferences(userId: string) {
  if (!browser) return;
  currentUserId = userId;
  contrastPreference.set(readStoredContrast(userId));
}

export function clearUserPreferences() {
  currentUserId = null;
  contrastPreference.set(DEFAULT_CONTRAST);
}
