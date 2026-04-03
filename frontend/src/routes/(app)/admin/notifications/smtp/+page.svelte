<script lang="ts">
  import { onMount } from 'svelte';
  import { BellRing, RefreshCw, Send, ShieldCheck } from '@lucide/svelte';
  import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';
  import { GetSMTPSettingsUseCase } from '$lib/application/useCases/notification/getSMTPSettingsUseCase.js';
  import { SaveSMTPSettingsUseCase } from '$lib/application/useCases/notification/saveSMTPSettingsUseCase.js';
  import { SendSMTPTestEmailUseCase } from '$lib/application/useCases/notification/sendSMTPTestEmailUseCase.js';
  import {
    SMTPOverviewCard,
    SMTPSettingsForm,
    SMTPTestEmailCard
  } from '$lib/components/notifications/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
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

<div class="flex flex-col gap-6">
  <section class="overflow-hidden rounded-2xl border bg-card">
    <div class="border-b bg-muted/30 px-6 py-5">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="space-y-3">
          <div
            class="inline-flex items-center gap-2 rounded-full border bg-background px-3 py-1 text-xs font-medium text-muted-foreground"
          >
            <ShieldCheck class="size-3.5 text-primary" />
            {$t('notifications.hero.scope')}
          </div>
          <div class="space-y-1">
            <h1 class="text-3xl font-semibold tracking-tight">{$t('notifications.page.title')}</h1>
            <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
              {$t('notifications.page.description')}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            onclick={() => loadSettings('refresh')}
            disabled={isLoading || isRefreshing || isSaving || isSendingTest}
          >
            <RefreshCw class={`size-4${isRefreshing ? ' animate-spin' : ''}`} />
            {$t('common.refresh')}
          </Button>
        </div>
      </div>
    </div>

    <div class="grid gap-3 px-6 py-5 md:grid-cols-3">
      <div class="rounded-xl border bg-background p-4">
        <p class="text-sm text-muted-foreground">{$t('notifications.hero.service_status')}</p>
        <div class="mt-2 flex items-center gap-2">
          <BellRing class="size-4 text-primary" />
          <p class="font-medium">{serviceStatus}</p>
        </div>
      </div>

      <div class="rounded-xl border bg-background p-4">
        <p class="text-sm text-muted-foreground">{$t('notifications.hero.delivery_channel')}</p>
        <div class="mt-2 flex items-center gap-2">
          <Send class="size-4 text-primary" />
          <p class="font-medium break-all">
            {settings
              ? `${settings.host}:${settings.port}`
              : $t('notifications.status.not_configured')}
          </p>
        </div>
      </div>

      <div class="rounded-xl border bg-background p-4">
        <p class="text-sm text-muted-foreground">{$t('notifications.hero.last_sync')}</p>
        <div class="mt-2 flex items-center gap-2">
          <RefreshCw class="size-4 text-primary" />
          <p class="font-medium">{formatDateTime(lastLoadedAt)}</p>
        </div>
      </div>
    </div>
  </section>

  {#if loadError}
    <Alert.Root variant="destructive">
      <Alert.Description>{loadError}</Alert.Description>
    </Alert.Root>
  {/if}

  <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_380px]">
    <SMTPSettingsForm
      {settings}
      hasStoredPassword={settings?.has_password ?? false}
      {isLoading}
      isSubmitting={isSaving}
      formError={saveError}
      fieldErrors={saveFieldErrors}
      onSubmit={handleSave}
    />

    <div class="flex flex-col gap-6">
      <SMTPOverviewCard {settings} {isLoading} />
      <SMTPTestEmailCard
        {settings}
        defaultRecipient={data.user.email}
        isSubmitting={isSendingTest}
        formError={testError}
        fieldErrors={testFieldErrors}
        onSubmit={handleSendTest}
      />
    </div>
  </div>
</div>
