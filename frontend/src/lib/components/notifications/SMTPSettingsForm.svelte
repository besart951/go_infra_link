<script lang="ts">
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Switch from '$lib/components/ui/switch/index.js';
  import { cn } from '$lib/utils.js';
  import type {
    SMTPAuthMode,
    SMTPSettings,
    SMTPSettingsFormValues,
    SMTPSecurityMode,
    UpsertSMTPSettingsRequest
  } from '$lib/domain/notification/index.js';
  import {
    createSMTPSettingsFormValues,
    getSMTPDefaultPort,
    isSMTPPasswordRequired,
    normalizeSMTPSettingsInput,
    validateSMTPSettingsInput
  } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import SMTPModeSelector from './SMTPModeSelector.svelte';

  interface Props {
    settings?: SMTPSettings | null;
    hasStoredPassword?: boolean;
    isLoading?: boolean;
    isSubmitting?: boolean;
    formError?: string | null;
    fieldErrors?: Record<string, string>;
    onSubmit?: (payload: UpsertSMTPSettingsRequest) => Promise<void> | void;
  }

  let {
    settings = null,
    hasStoredPassword = false,
    isLoading = false,
    isSubmitting = false,
    formError = null,
    fieldErrors = {},
    onSubmit
  }: Props = $props();

  const t = createTranslator();

  let draft = $state<SMTPSettingsFormValues>(createSMTPSettingsFormValues(null));
  let localErrors = $state<Record<string, string>>({});
  let loadedSignature = $state('');

  $effect(() => {
    const nextSignature = settings
      ? `${settings.id}:${settings.updated_at}:${settings.has_password}`
      : `empty:${hasStoredPassword}`;

    if (nextSignature !== loadedSignature) {
      draft = createSMTPSettingsFormValues(settings);
      localErrors = {};
      loadedSignature = nextSignature;
    }
  });

  const isDirty = $derived.by(() => {
    const initial = normalizeSMTPSettingsInput(createSMTPSettingsFormValues(settings));
    const current = normalizeSMTPSettingsInput(draft);
    return JSON.stringify(initial) !== JSON.stringify(current);
  });

  const requiresCredentials = $derived(draft.auth_mode === 'plain');
  const passwordRequired = $derived(isSMTPPasswordRequired(draft.auth_mode, hasStoredPassword));

  const securityOptions = $derived([
    {
      value: 'none',
      label: $t('notifications.security.none'),
      description: $t('notifications.security_descriptions.none'),
      badge: '25'
    },
    {
      value: 'starttls',
      label: $t('notifications.security.starttls'),
      description: $t('notifications.security_descriptions.starttls'),
      badge: '587'
    },
    {
      value: 'tls',
      label: $t('notifications.security.tls'),
      description: $t('notifications.security_descriptions.tls'),
      badge: '465'
    }
  ]);

  const authOptions = $derived([
    {
      value: 'none',
      label: $t('notifications.auth.none'),
      description: $t('notifications.auth_descriptions.none')
    },
    {
      value: 'plain',
      label: $t('notifications.auth.plain'),
      description: $t('notifications.auth_descriptions.plain')
    }
  ]);

  function errorFor(field: string): string | undefined {
    const key = fieldErrors[field] ?? localErrors[field];
    return key ? $t(key) : undefined;
  }

  function selectSecurity(value: string) {
    const next = value as SMTPSecurityMode;
    const currentPort = getSMTPDefaultPort(draft.security);
    const nextPort = getSMTPDefaultPort(next);

    draft.security = next;

    if (!draft.port || draft.port === currentPort) {
      draft.port = nextPort;
    }

    if (next === 'none') {
      draft.allow_insecure_tls = false;
    }
  }

  function selectAuthMode(value: string) {
    draft.auth_mode = value as SMTPAuthMode;
  }

  function resetDraft() {
    draft = createSMTPSettingsFormValues(settings);
    localErrors = {};
  }

  function toggleEnabled() {
    if (isLoading || isSubmitting) return;
    draft.enabled = !draft.enabled;
  }

  function handleEnabledCardClick(event: MouseEvent) {
    if ((event.target as HTMLElement | null)?.closest('[data-slot="switch"]')) return;
    toggleEnabled();
  }

  function handleEnabledCardKeydown(event: KeyboardEvent) {
    if (event.key !== ' ' && event.key !== 'Enter') return;
    event.preventDefault();
    toggleEnabled();
  }

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();

    const payload = normalizeSMTPSettingsInput(draft);
    localErrors = validateSMTPSettingsInput(payload, hasStoredPassword);

    if (Object.keys(localErrors).length > 0) {
      return;
    }

    await onSubmit?.(payload);
  }
</script>

<Card.Root class="overflow-hidden rounded-lg shadow-none">
  <Card.Header class="gap-3 border-b px-4 py-4 sm:px-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div class="space-y-1">
        <Card.Title class="text-xl leading-7">{$t('notifications.form.title')}</Card.Title>
        <Card.Description class="leading-6">{$t('notifications.form.description')}</Card.Description
        >
      </div>
      <Badge
        class="w-fit shrink-0"
        variant={isLoading ? 'secondary' : isDirty ? 'warning' : 'secondary'}
      >
        {$t(
          isLoading
            ? 'notifications.status.loading'
            : isDirty
              ? 'notifications.status.unsaved'
              : 'notifications.status.synced'
        )}
      </Badge>
    </div>
  </Card.Header>

  <Card.Content class="px-0">
    <form class="divide-y" onsubmit={handleSubmit}>
      {#if formError}
        <div class="px-4 py-4 sm:px-6">
          <Alert.Root variant="destructive">
            <Alert.Description>{formError}</Alert.Description>
          </Alert.Root>
        </div>
      {/if}

      <section class="px-4 py-5 sm:px-6">
        <div
          role="button"
          tabindex={isLoading || isSubmitting ? -1 : 0}
          aria-labelledby="smtp_enabled_title"
          aria-describedby="smtp_enabled_description"
          aria-disabled={isLoading || isSubmitting}
          class={cn(
            'flex w-full cursor-pointer flex-col gap-4 rounded-lg border bg-muted/20 p-4 text-left transition-colors outline-none hover:bg-muted/35 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 sm:flex-row sm:items-center sm:justify-between',
            draft.enabled && 'border-primary/60 bg-primary/5 hover:bg-primary/10',
            (isLoading || isSubmitting) && 'cursor-not-allowed opacity-60'
          )}
          onclick={handleEnabledCardClick}
          onkeydown={handleEnabledCardKeydown}
        >
          <span class="min-w-0 space-y-1">
            <span id="smtp_enabled_title" class="block text-sm font-medium">
              {$t('notifications.form.enable_title')}
            </span>
            <span
              id="smtp_enabled_description"
              class="block text-sm leading-6 text-muted-foreground"
            >
              {$t('notifications.form.enable_description')}
            </span>
          </span>
          <span class="flex items-center gap-3">
            <span class="text-sm font-medium whitespace-nowrap">
              {$t(draft.enabled ? 'notifications.status.enabled' : 'notifications.status.disabled')}
            </span>
            <Switch.Root
              id="smtp_enabled"
              bind:checked={draft.enabled}
              disabled={isLoading || isSubmitting}
              size="lg"
              aria-labelledby="smtp_enabled_title"
              aria-describedby="smtp_enabled_description"
              class="shadow-sm shadow-primary/20 data-unchecked:bg-muted"
              onclick={(event) => event.stopPropagation()}
            />
          </span>
        </div>
      </section>

      <section class="grid gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[12rem_minmax(0,1fr)]">
        <div class="flex items-start gap-3">
          <span
            class="flex size-7 shrink-0 items-center justify-center rounded-full border bg-background text-xs font-semibold"
          >
            1
          </span>
          <div class="min-w-0 space-y-1">
            <p class="text-sm font-semibold">{$t('notifications.setup.server_step')}</p>
            <p class="text-sm leading-6 text-muted-foreground">
              {$t('notifications.setup.server_description')}
            </p>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-[minmax(0,1fr)_8rem]">
          <div class="space-y-2">
            <Label for="smtp_host">{$t('notifications.form.host')}</Label>
            <Input
              id="smtp_host"
              bind:value={draft.host}
              disabled={isLoading || isSubmitting}
              placeholder={$t('notifications.form.host_placeholder')}
              aria-invalid={Boolean(errorFor('host'))}
            />
            {#if errorFor('host')}
              <p class="text-sm text-destructive">{errorFor('host')}</p>
            {/if}
          </div>

          <div class="space-y-2">
            <Label for="smtp_port">{$t('notifications.form.port')}</Label>
            <Input
              id="smtp_port"
              type="number"
              bind:value={draft.port}
              disabled={isLoading || isSubmitting}
              min={1}
              max={65535}
              aria-invalid={Boolean(errorFor('port'))}
            />
            {#if errorFor('port')}
              <p class="text-sm text-destructive">{errorFor('port')}</p>
            {/if}
          </div>
        </div>
      </section>

      <section class="grid gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[12rem_minmax(0,1fr)]">
        <div class="flex items-start gap-3">
          <span
            class="flex size-7 shrink-0 items-center justify-center rounded-full border bg-background text-xs font-semibold"
          >
            2
          </span>
          <div class="min-w-0 space-y-1">
            <p class="text-sm font-semibold">{$t('notifications.setup.sender_step')}</p>
            <p class="text-sm leading-6 text-muted-foreground">
              {$t('notifications.setup.sender_description')}
            </p>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="smtp_from_email">{$t('notifications.form.from_email')}</Label>
            <Input
              id="smtp_from_email"
              type="email"
              bind:value={draft.from_email}
              disabled={isLoading || isSubmitting}
              placeholder={$t('notifications.form.from_email_placeholder')}
              aria-invalid={Boolean(errorFor('from_email'))}
            />
            {#if errorFor('from_email')}
              <p class="text-sm text-destructive">{errorFor('from_email')}</p>
            {/if}
          </div>

          <div class="space-y-2">
            <Label for="smtp_from_name">{$t('notifications.form.from_name')}</Label>
            <Input
              id="smtp_from_name"
              bind:value={draft.from_name}
              disabled={isLoading || isSubmitting}
              placeholder={$t('notifications.form.from_name_placeholder')}
            />
          </div>

          <div class="space-y-2 sm:col-span-2">
            <Label for="smtp_reply_to">
              {$t('notifications.form.reply_to')}
              <span class="ms-2 text-xs font-normal text-muted-foreground">
                {$t('pages.optional')}
              </span>
            </Label>
            <Input
              id="smtp_reply_to"
              type="email"
              bind:value={draft.reply_to}
              disabled={isLoading || isSubmitting}
              placeholder={$t('notifications.form.reply_to_placeholder')}
              aria-invalid={Boolean(errorFor('reply_to'))}
            />
            {#if errorFor('reply_to')}
              <p class="text-sm text-destructive">{errorFor('reply_to')}</p>
            {/if}
          </div>
        </div>
      </section>

      <section class="grid gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[12rem_minmax(0,1fr)]">
        <div class="flex items-start gap-3">
          <span
            class="flex size-7 shrink-0 items-center justify-center rounded-full border bg-background text-xs font-semibold"
          >
            3
          </span>
          <div class="min-w-0 space-y-1">
            <p class="text-sm font-semibold">{$t('notifications.setup.security_step')}</p>
            <p class="text-sm leading-6 text-muted-foreground">
              {$t('notifications.setup.security_description')}
            </p>
          </div>
        </div>

        <div class="space-y-5">
          <SMTPModeSelector
            value={draft.security}
            label={$t('notifications.form.transport_title')}
            description={$t('notifications.form.transport_description')}
            options={securityOptions}
            columnsClass="xl:grid-cols-3"
            onChange={selectSecurity}
          />

          {#if draft.security !== 'none'}
            <div class="flex items-start gap-3 rounded-lg border bg-background px-4 py-3">
              <Switch.Root
                id="smtp_allow_insecure_tls"
                bind:checked={draft.allow_insecure_tls}
                disabled={isLoading || isSubmitting}
                class="mt-1"
              />
              <span class="min-w-0 space-y-1">
                <Label for="smtp_allow_insecure_tls" class="block text-sm font-medium">
                  {$t('notifications.form.allow_insecure_title')}
                </Label>
                <span class="block text-sm leading-6 text-muted-foreground">
                  {$t('notifications.form.allow_insecure_description')}
                </span>
              </span>
            </div>
          {/if}
        </div>
      </section>

      <section class="grid gap-5 px-4 py-5 sm:px-6 lg:grid-cols-[12rem_minmax(0,1fr)]">
        <div class="flex items-start gap-3">
          <span
            class="flex size-7 shrink-0 items-center justify-center rounded-full border bg-background text-xs font-semibold"
          >
            4
          </span>
          <div class="min-w-0 space-y-1">
            <p class="text-sm font-semibold">{$t('notifications.setup.auth_step')}</p>
            <p class="text-sm leading-6 text-muted-foreground">
              {$t('notifications.setup.auth_description')}
            </p>
          </div>
        </div>

        <div class="space-y-5">
          <SMTPModeSelector
            value={draft.auth_mode}
            label={$t('notifications.form.auth_title')}
            description={$t('notifications.form.auth_description')}
            options={authOptions}
            columnsClass="lg:grid-cols-2"
            onChange={selectAuthMode}
          />

          {#if requiresCredentials}
            <div class="grid gap-4 sm:grid-cols-2">
              <div class="space-y-2">
                <Label for="smtp_username">{$t('notifications.form.username')}</Label>
                <Input
                  id="smtp_username"
                  bind:value={draft.username}
                  disabled={isLoading || isSubmitting}
                  placeholder={$t('notifications.form.username_placeholder')}
                  aria-invalid={Boolean(errorFor('username'))}
                />
                {#if errorFor('username')}
                  <p class="text-sm text-destructive">{errorFor('username')}</p>
                {/if}
              </div>

              <div class="space-y-2">
                <Label for="smtp_password">
                  {$t('notifications.form.password')}
                  {#if !passwordRequired}
                    <span class="ms-2 text-xs font-normal text-muted-foreground">
                      {$t('pages.optional')}
                    </span>
                  {/if}
                </Label>
                <Input
                  id="smtp_password"
                  type="password"
                  bind:value={draft.password}
                  disabled={isLoading || isSubmitting}
                  placeholder={$t('notifications.form.password_placeholder')}
                  aria-invalid={Boolean(errorFor('password'))}
                />
                {#if errorFor('password')}
                  <p class="text-sm text-destructive">{errorFor('password')}</p>
                {:else if passwordRequired}
                  <p class="text-sm text-muted-foreground">
                    {$t('notifications.form.password_required')}
                  </p>
                {:else if hasStoredPassword}
                  <p class="text-sm text-muted-foreground">
                    {$t('notifications.form.password_kept')}
                  </p>
                {/if}
              </div>
            </div>
          {/if}
        </div>
      </section>

      <div
        class="flex flex-col gap-3 bg-muted/20 px-4 py-4 sm:px-6 md:flex-row md:items-center md:justify-between"
      >
        <p class="text-sm leading-6 text-muted-foreground">
          {$t('notifications.form.description')}
        </p>

        <div class="flex flex-col gap-2 sm:flex-row">
          <Button
            type="button"
            variant="outline"
            class="w-full sm:w-auto"
            onclick={resetDraft}
            disabled={isLoading || isSubmitting || !isDirty}
          >
            {$t('notifications.form.reset')}
          </Button>
          <Button type="submit" class="w-full sm:w-auto" disabled={isLoading || isSubmitting}>
            {$t(isSubmitting ? 'common.saving' : 'notifications.form.save')}
          </Button>
        </div>
      </div>
    </form>
  </Card.Content>
</Card.Root>
