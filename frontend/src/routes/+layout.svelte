<script lang="ts">
  import '@fontsource-variable/noto-sans';
  import { onMount } from 'svelte';
  import './layout.css';
  import { initAppearance } from '$lib/stores/appearance.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { i18n } from '$lib/i18n/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  const { children } = $props();
  const t = createTranslator();

  let translationsReady = $state(false);
  let translationError = $state<string | null>(null);

  onMount(() => {
    initAppearance();

    // Subscribe to i18n store to know when translations are loaded
    const unsubscribe = i18n.subscribe((state) => {
      translationsReady = !state.isLoading;
      translationError = state.error;
    });

    return unsubscribe;
  });
</script>

{#if !translationsReady && !translationError}
  <div class="flex h-screen items-center justify-center">
    <div class="text-center">
      <div
        class="mb-4 inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]"
      ></div>
      <p class="text-muted-foreground">{$t('app.loading')}</p>
    </div>
  </div>
{:else if translationError}
  <div class="flex h-screen items-center justify-center">
    <div class="text-center text-destructive">
      <p class="mb-2 text-lg font-semibold">{$t('app.translation_error_title')}</p>
      <p class="text-sm">{translationError}</p>
      <Button onclick={() => i18n.reload()} class="mt-4">
        {$t('app.retry')}
      </Button>
    </div>
  </div>
{:else}
  {@render children()}
{/if}
