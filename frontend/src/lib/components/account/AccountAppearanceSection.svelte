<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator';
  import {
    CONTRAST_PREFERENCE_STEP,
    DEFAULT_CONTRAST_PREFERENCE,
    FONT_STACKS,
    MAX_CONTRAST_PREFERENCE,
    MIN_CONTRAST_PREFERENCE,
    contrastPreference,
    fontPreference,
    setContrastPreference,
    setFontPreference,
    setThemePreference,
    themePreference,
    type FontPreference,
    type ThemePreference
  } from '$lib/stores/appearance.js';
  import { Contrast, LaptopMinimal, Moon, RotateCcw, Sun, Type } from '@lucide/svelte';
  import type { Component } from 'svelte';

  type ThemeOption = {
    value: ThemePreference;
    label: string;
    description: string;
    icon: Component;
  };

  type FontOption = {
    value: FontPreference;
    label: string;
    description: string;
    stack: string;
  };

  const t = createTranslator();

  const themeOptions = $derived<ThemeOption[]>([
    {
      value: 'system',
      label: $t('pages.settings_theme_system'),
      description: $t('pages.settings_theme_system_desc'),
      icon: LaptopMinimal
    },
    {
      value: 'light',
      label: $t('pages.settings_theme_light'),
      description: $t('pages.settings_theme_light_desc'),
      icon: Sun
    },
    {
      value: 'dark',
      label: $t('pages.settings_theme_dark'),
      description: $t('pages.settings_theme_dark_desc'),
      icon: Moon
    }
  ]);

  const fontOptions = $derived<FontOption[]>([
    {
      value: 'noto',
      label: $t('pages.settings_font_noto'),
      description: $t('pages.settings_font_noto_desc'),
      stack: FONT_STACKS.noto
    },
    {
      value: 'system',
      label: $t('pages.settings_font_system'),
      description: $t('pages.settings_font_system_desc'),
      stack: FONT_STACKS.system
    },
    {
      value: 'serif',
      label: $t('pages.settings_font_serif'),
      description: $t('pages.settings_font_serif_desc'),
      stack: FONT_STACKS.serif
    },
    {
      value: 'mono',
      label: $t('pages.settings_font_mono'),
      description: $t('pages.settings_font_mono_desc'),
      stack: FONT_STACKS.mono
    }
  ]);

  function handleContrastInput(event: Event) {
    setContrastPreference(Number((event.currentTarget as HTMLInputElement).value));
  }
</script>

<div class="rounded-lg border bg-card p-4">
  <div class="flex flex-col gap-1">
    <h2 class="text-base font-semibold">{$t('pages.settings_appearance')}</h2>
    <p class="text-sm text-muted-foreground">{$t('pages.settings_appearance_desc')}</p>
  </div>

  <section class="mt-4 grid gap-2 sm:grid-cols-3">
    {#each themeOptions as opt (opt.value)}
      {@const active = $themePreference === opt.value}
      <Button
        variant={active ? 'default' : 'outline'}
        class="h-full items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
        aria-pressed={active}
        onclick={() => setThemePreference(opt.value)}
      >
        <opt.icon class="mt-0.5 size-4 shrink-0" />
        <span class="flex min-w-0 flex-col items-start gap-0.5 text-left">
          <span class="leading-tight">{opt.label}</span>
          <span
            class={active
              ? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
              : 'text-xs leading-snug wrap-break-word text-muted-foreground'}
          >
            {opt.description}
          </span>
        </span>
      </Button>
    {/each}
  </section>

  <section class="mt-6 border-t pt-5">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div class="flex min-w-0 items-start gap-2">
        <Contrast class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
        <div class="min-w-0 space-y-1">
          <h3 class="text-sm font-semibold">{$t('pages.settings_contrast')}</h3>
          <p class="text-sm text-muted-foreground">{$t('pages.settings_contrast_desc')}</p>
        </div>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <span class="rounded-md bg-muted px-2 py-1 text-sm font-medium tabular-nums">
          {$contrastPreference}%
        </span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={$contrastPreference === DEFAULT_CONTRAST_PREFERENCE}
          onclick={() => setContrastPreference(DEFAULT_CONTRAST_PREFERENCE)}
        >
          <RotateCcw class="size-4" />
          {$t('pages.settings_contrast_reset')}
        </Button>
      </div>
    </div>

    <div class="mt-4">
      <label for="appearance_contrast" class="sr-only">
        {$t('pages.settings_contrast')}
      </label>
      <input
        id="appearance_contrast"
        type="range"
        min={MIN_CONTRAST_PREFERENCE}
        max={MAX_CONTRAST_PREFERENCE}
        step={CONTRAST_PREFERENCE_STEP}
        value={$contrastPreference}
        class="h-2 w-full cursor-pointer appearance-none rounded-full bg-muted accent-primary"
        oninput={handleContrastInput}
      />
      <div class="mt-2 flex justify-between text-xs text-muted-foreground">
        <span>{$t('pages.settings_contrast_min')}</span>
        <span>{$t('pages.settings_contrast_default')}</span>
        <span>{$t('pages.settings_contrast_max')}</span>
      </div>
    </div>
  </section>

  <section class="mt-6 border-t pt-5">
    <div class="flex min-w-0 items-start gap-2">
      <Type class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
      <div class="min-w-0 space-y-1">
        <h3 class="text-sm font-semibold">{$t('pages.settings_font')}</h3>
        <p class="text-sm text-muted-foreground">{$t('pages.settings_font_desc')}</p>
      </div>
    </div>

    <div class="mt-4 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
      {#each fontOptions as font (font.value)}
        {@const active = $fontPreference === font.value}
        <Button
          type="button"
          variant={active ? 'default' : 'outline'}
          class="h-full min-h-28 items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
          aria-pressed={active}
          onclick={() => setFontPreference(font.value)}
        >
          <Type class="mt-0.5 size-4 shrink-0" />
          <span class="flex min-w-0 flex-col items-start gap-1 text-left">
            <span class="leading-tight" style={`font-family: ${font.stack}`}>
              {font.label}
            </span>
            <span
              class={active
                ? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
                : 'text-xs leading-snug wrap-break-word text-muted-foreground'}
            >
              {font.description}
            </span>
            <span
              class={active
                ? 'font-mono text-[11px] leading-snug wrap-break-word text-primary-foreground/70'
                : 'font-mono text-[11px] leading-snug wrap-break-word text-muted-foreground'}
            >
              {font.stack}
            </span>
          </span>
        </Button>
      {/each}
    </div>
  </section>
</div>
