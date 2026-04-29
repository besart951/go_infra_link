<script lang="ts">
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { ArrowLeft, FileQuestion, Home, ShieldAlert, TriangleAlert } from '@lucide/svelte';

  interface Props {
    status?: number;
    from?: string | null;
    fallbackMessage?: string;
  }

  let { status = 500, from = null, fallbackMessage }: Props = $props();

  const t = createTranslator();
  const normalizedStatus = $derived(status === 403 || status === 404 ? status : 500);
  const titleKey = $derived(
    normalizedStatus === 403
      ? 'pages.http_error.forbidden.title'
      : normalizedStatus === 404
        ? 'pages.http_error.not_found.title'
        : 'pages.http_error.generic.title'
  );
  const descriptionKey = $derived(
    normalizedStatus === 403
      ? 'pages.http_error.forbidden.description'
      : normalizedStatus === 404
        ? 'pages.http_error.not_found.description'
        : 'pages.http_error.generic.description'
  );
  const eyebrowKey = $derived(
    normalizedStatus === 403
      ? 'pages.http_error.forbidden.eyebrow'
      : normalizedStatus === 404
        ? 'pages.http_error.not_found.eyebrow'
        : 'pages.http_error.generic.eyebrow'
  );

  function goBack() {
    if (browser && window.history.length > 1) {
      window.history.back();
      return;
    }

    void goto(from || '/');
  }
</script>

<svelte:head>
  <title>{$t(titleKey)} | Infra Link</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4 py-10">
  <section class="w-full max-w-xl rounded-lg border bg-background p-6 shadow-sm">
    <div class="mb-5 flex items-center gap-3">
      <div
        class="flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground"
      >
        {#if normalizedStatus === 403}
          <ShieldAlert class="h-6 w-6" />
        {:else if normalizedStatus === 404}
          <FileQuestion class="h-6 w-6" />
        {:else}
          <TriangleAlert class="h-6 w-6" />
        {/if}
      </div>
      <div class="min-w-0">
        <p class="text-sm font-medium text-muted-foreground">{$t(eyebrowKey)}</p>
        <h1 class="text-2xl font-semibold tracking-normal text-foreground">{$t(titleKey)}</h1>
      </div>
    </div>

    <p class="text-sm leading-6 text-muted-foreground">
      {fallbackMessage || $t(descriptionKey)}
    </p>

    <div class="mt-6 flex flex-wrap gap-2">
      <Button onclick={goBack}>
        <ArrowLeft class="h-4 w-4" />
        {$t('pages.http_error.back')}
      </Button>
      <Button variant="outline" href="/">
        <Home class="h-4 w-4" />
        {$t('pages.http_error.dashboard')}
      </Button>
    </div>
  </section>
</div>
