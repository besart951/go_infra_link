import { get } from 'svelte/store';
import { beforeEach, describe, expect, it, vi } from 'vitest';

async function loadAppearanceStore() {
  vi.resetModules();
  return import('./appearance');
}

describe('appearance preferences', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.style.removeProperty('--app-contrast');
    document.documentElement.style.removeProperty('--app-contrast-filter');
  });

  it('keeps contrast at 100 for a new user without a stored preference', async () => {
    const { contrastPreference, initAppearance } = await loadAppearanceStore();

    initAppearance('user-1');

    expect(get(contrastPreference)).toBe(100);
    expect(localStorage.getItem('contrast_preference:user-1')).toBe('100');
    expect(document.documentElement.style.getPropertyValue('--app-contrast')).toBe('100%');
    expect(document.documentElement.style.getPropertyValue('--app-contrast-filter')).toBe('none');
  });

  it('does not treat a missing contrast value as the minimum', async () => {
    const { contrastPreference, initAppearance } = await loadAppearanceStore();

    initAppearance();

    expect(get(contrastPreference)).toBe(100);
    expect(localStorage.getItem('contrast_preference')).toBe('100');
  });
});
