<script lang="ts">
  import { onMount } from 'svelte';
  import { MailCheck, RefreshCw, ServerCog, ShieldCheck } from '@lucide/svelte';
  import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';
  import { GetSMTPSettingsUseCase } from '$lib/application/useCases/notification/getSMTPSettingsUseCase.js';
  import { SaveSMTPSettingsUseCase } from '$lib/application/useCases/notification/saveSMTPSettingsUseCase.js';
  import { SendSMTPTestEmailUseCase } from '$lib/application/useCases/notification/sendSMTPTestEmailUseCase.js';
  import {
    NotificationRulesCard,
    SMTPOverviewCard,
    SMTPSettingsForm,
    SMTPTestEmailCard
  } from '$lib/components/notifications/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import type {
    SendSMTPTestEmailRequest,
    SMTPSettings,
    UpsertSMTPSettingsRequest
  } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { notificationSettingsRepository } from '$lib/infrastructure/api/notificationSettingsRepository.js';
  import type { PageData } from './$types.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();

  const getSMTPSettings = new GetSMTPSettingsUseCase(notificationSettingsRepository);
  const saveSMTPSettings = new SaveSMTPSettingsUseCase(notificationSettingsRepository);
  const sendSMTPTestEmail = new SendSMTPTestEmailUseCase(notificationSettingsRepository);

  let settings = $state<SMTPSettings | null>(null);
  let isLoading = $state(true);
  let isRefreshing = $state(false);
  let loadError = $state<string | null>(null);
  let lastLoadedAt = $state<string | null>(null);

  let isSaving = $state(false);
  let saveError = $state<string | null>(null);
  let saveFieldErrors = $state<Record<string, string>>({});

  let isSendingTest = $state(false);
  let testError = $state<string | null>(null);
  let testFieldErrors = $state<Record<string, string>>({});

  const serviceStatus = $derived.by(() => {
    if (!settings) return $t('notifications.status.not_configured');
    return $t(settings.enabled ? 'notifications.status.enabled' : 'notifications.status.disabled');
  });

  const connectionLabel = $derived(
    settings ? `${settings.host}:${settings.port}` : $t('notifications.status.not_configured')
  );

  function statusVariant(): 'secondary' | 'outline' | 'success' | 'warning' {
    if (isLoading) return 'secondary';
    if (!settings) return 'outline';
    return settings.enabled ? 'success' : 'warning';
  }

  async function loadSettings(mode: 'initial' | 'refresh' = 'initial') {
    if (mode === 'initial') {
      isLoading = true;
    } else {
      isRefreshing = true;
    }

    loadError = null;

    try {
      settings = await getSMTPSettings.execute();
      lastLoadedAt = new Date().toISOString();

      if (mode === 'refresh') {
        addToast($t('notifications.toasts.refreshed'), 'success');
      }
    } catch (error) {
      loadError = getErrorMessage(error);

      if (mode === 'refresh') {
        addToast(loadError, 'error');
      }
    } finally {
      isLoading = false;
      isRefreshing = false;
    }
  }

  async function handleSave(payload: UpsertSMTPSettingsRequest) {
    isSaving = true;
    saveError = null;
    saveFieldErrors = {};

    try {
      settings = await saveSMTPSettings.execute(payload);
      lastLoadedAt = new Date().toISOString();
      saveError = null;
      saveFieldErrors = {};
      addToast($t('notifications.toasts.saved'), 'success');
    } catch (error) {
      saveFieldErrors = getFieldErrors(error);
      saveError = Object.keys(saveFieldErrors).length === 0 ? getErrorMessage(error) : null;

      if (saveError) {
        addToast(saveError, 'error');
      }
    } finally {
      isSaving = false;
    }
  }

  async function handleSendTest(payload: SendSMTPTestEmailRequest) {
    isSendingTest = true;
    testError = null;
    testFieldErrors = {};

    try {
      await sendSMTPTestEmail.execute(payload);
      testError = null;
      testFieldErrors = {};
      addToast($t('notifications.toasts.test_sent'), 'success');
    } catch (error) {
      testFieldErrors = getFieldErrors(error);
      testError = Object.keys(testFieldErrors).length === 0 ? getErrorMessage(error) : null;

      if (testError) {
        addToast(testError, 'error');
      }
    } finally {
      isSendingTest = false;
    }
  }

  function formatDateTime(value: string | null): string {
    if (!value) return $t('common.not_available');
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  onMount(() => {
    loadSettings();
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
        <Badge variant={statusVariant()} class="gap-1.5">
          <MailCheck class="size-3.5" />
          {serviceStatus}
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
          <p class="truncate font-medium">{connectionLabel}</p>
        </div>
      </div>

      <div
        class="flex min-w-0 items-center gap-2 rounded-lg border bg-background px-3 py-2 text-sm"
      >
        <RefreshCw class="size-4 shrink-0 text-muted-foreground" />
        <div class="min-w-0">
          <p class="text-xs text-muted-foreground">{$t('notifications.hero.last_sync')}</p>
          <p class="truncate font-medium">{formatDateTime(lastLoadedAt)}</p>
        </div>
      </div>

      <Button
        variant="outline"
        class="w-full sm:w-auto"
        onclick={() => loadSettings('refresh')}
        disabled={isLoading || isRefreshing || isSaving || isSendingTest}
      >
        <RefreshCw class={`size-4${isRefreshing ? ' animate-spin' : ''}`} />
        {$t('common.refresh')}
      </Button>
    </div>
  </header>

  {#if loadError}
    <Alert.Root variant="destructive">
      <Alert.Description>{loadError}</Alert.Description>
    </Alert.Root>
  {/if}

  <div class="grid items-start gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(20rem,24rem)]">
    <SMTPSettingsForm
      {settings}
      hasStoredPassword={settings?.has_password ?? false}
      {isLoading}
      isSubmitting={isSaving}
      formError={saveError}
      fieldErrors={saveFieldErrors}
      onSubmit={handleSave}
    />

    <aside class="flex min-w-0 flex-col gap-5 xl:sticky xl:top-20">
      <SMTPOverviewCard {settings} {isLoading} />
      <SMTPTestEmailCard
        {settings}
        defaultRecipient={data.user.email}
        isSubmitting={isSendingTest}
        formError={testError}
        fieldErrors={testFieldErrors}
        onSubmit={handleSendTest}
      />
    </aside>
  </div>

  <NotificationRulesCard />
</div>
