<script lang="ts">
  import { onMount } from 'svelte';
  import { MailCheck, RefreshCw, ShieldCheck } from '@lucide/svelte';
  import {
    NotificationRulesCard,
    SMTPOverviewCard,
    SMTPSettingsForm,
    SMTPTestEmailCard
  } from '$lib/components/notifications/index.js';
  import { SMTPManagementPageState } from '$lib/components/notifications/SMTPManagementPageState.svelte.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { buttonVariants } from '$lib/components/ui/button/index.js';
  import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
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
  <EntityListHeader
    title={$t('notifications.page.title')}
    description={$t('notifications.page.description')}
    backHref="/notifications"
    backLabel={$t('hub.back_to_overview')}
  >
    <Badge variant="outline" class="hidden gap-1.5 sm:inline-flex">
      <ShieldCheck class="size-3.5" />
      {$t('notifications.hero.scope')}
    </Badge>
    <Badge variant={state.statusVariant()} class="gap-1.5">
      <MailCheck class="size-3.5" />
      {state.serviceStatus}
    </Badge>
    <ButtonGroup.Root>
      <Tooltip.Root>
        <Tooltip.Trigger
          class={buttonVariants({ variant: 'outline', size: 'icon' })}
          onclick={() => state.loadSettings('refresh')}
          disabled={state.isLoading || state.isRefreshing || state.isSaving || state.isSendingTest}
          aria-label={$t('common.refresh')}
        >
          <RefreshCw class={`size-4${state.isRefreshing ? ' animate-spin' : ''}`} />
        </Tooltip.Trigger>
        <Tooltip.Content>{$t('common.refresh')}</Tooltip.Content>
      </Tooltip.Root>
    </ButtonGroup.Root>
  </EntityListHeader>

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
