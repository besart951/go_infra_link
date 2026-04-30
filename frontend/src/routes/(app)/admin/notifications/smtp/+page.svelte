<script lang="ts">
  import { onMount } from 'svelte';
  import { MailCheck, RefreshCw, ServerCog, ShieldCheck } from '@lucide/svelte';
  import {
    NotificationRulesCard,
    SMTPOverviewCard,
    SMTPSettingsForm,
    SMTPTestEmailCard
  } from '$lib/components/notifications/index.js';
  import { SMTPManagementPageState } from '$lib/components/notifications/SMTPManagementPageState.svelte.js';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { PageData } from './$types.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const state = new SMTPManagementPageState();

  onMount(() => {
    void state.loadSettings();
  });
</script>

<svelte:head>
  <title>{$t('notifications.page.title')} | {$t('app.brand')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-7xl flex-col gap-5">
  <header class="flex flex-col gap-4 border-b pb-5 lg:flex-row lg:items-end lg:justify-between">
    <div class="min-w-0 space-y-3">
      <div class="flex flex-wrap items-center gap-2">
        <Badge variant="outline" class="gap-1.5">
          <ShieldCheck class="size-3.5" />
          {$t('notifications.hero.scope')}
        </Badge>
        <Badge variant={state.statusVariant()} class="gap-1.5">
          <MailCheck class="size-3.5" />
          {state.serviceStatus}
        </Badge>
      </div>

      <div class="space-y-1">
        <h1 class="text-2xl leading-8 font-semibold tracking-tight sm:text-3xl">
          {$t('notifications.page.title')}
        </h1>
        <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
          {$t('notifications.page.description')}
        </p>
      </div>
    </div>

    <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
      <div
        class="flex min-w-0 items-center gap-2 rounded-lg border bg-background px-3 py-2 text-sm"
      >
        <ServerCog class="size-4 shrink-0 text-muted-foreground" />
        <div class="min-w-0">
          <p class="text-xs text-muted-foreground">{$t('notifications.hero.delivery_channel')}</p>
          <p class="truncate font-medium">{state.connectionLabel}</p>
        </div>
      </div>

      <div
        class="flex min-w-0 items-center gap-2 rounded-lg border bg-background px-3 py-2 text-sm"
      >
        <RefreshCw class="size-4 shrink-0 text-muted-foreground" />
        <div class="min-w-0">
          <p class="text-xs text-muted-foreground">{$t('notifications.hero.last_sync')}</p>
          <p class="truncate font-medium">{state.formatDateTime(state.lastLoadedAt)}</p>
        </div>
      </div>

      <Button
        variant="outline"
        class="w-full sm:w-auto"
        onclick={() => state.loadSettings('refresh')}
        disabled={state.isLoading || state.isRefreshing || state.isSaving || state.isSendingTest}
      >
        <RefreshCw class={`size-4${state.isRefreshing ? ' animate-spin' : ''}`} />
        {$t('common.refresh')}
      </Button>
    </div>
  </header>

  {#if state.loadError}
    <Alert.Root variant="destructive">
      <Alert.Description>{state.loadError}</Alert.Description>
    </Alert.Root>
  {/if}

  <div class="grid items-start gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(20rem,24rem)]">
    <SMTPSettingsForm
      settings={state.settings}
      hasStoredPassword={state.settings?.has_password ?? false}
      isLoading={state.isLoading}
      isSubmitting={state.isSaving}
      formError={state.saveError}
      fieldErrors={state.saveFieldErrors}
      onSubmit={(payload) => state.handleSave(payload)}
    />

    <aside class="flex min-w-0 flex-col gap-5 xl:sticky xl:top-20">
      <SMTPOverviewCard settings={state.settings} isLoading={state.isLoading} />
      <SMTPTestEmailCard
        settings={state.settings}
        defaultRecipient={data.user.email}
        isSubmitting={state.isSendingTest}
        formError={state.testError}
        fieldErrors={state.testFieldErrors}
        onSubmit={(payload) => state.handleSendTest(payload)}
      />
    </aside>
  </div>

  <NotificationRulesCard />
</div>
