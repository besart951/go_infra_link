<script lang="ts">
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
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

<Card.Root>
  <Card.Header class="gap-3">
    <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
      <div class="space-y-1">
        <Card.Title>{$t('notifications.form.title')}</Card.Title>
        <Card.Description>{$t('notifications.form.description')}</Card.Description>
      </div>
      <Badge variant={isLoading ? 'secondary' : isDirty ? 'outline' : 'secondary'}>
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

  <Card.Content>
    <form class="space-y-6" onsubmit={handleSubmit}>
      {#if formError}
        <Alert.Root variant="destructive">
          <Alert.Description>{formError}</Alert.Description>
        </Alert.Root>
      {/if}

      <div class="rounded-xl border bg-muted/20 p-4">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-1">
            <p class="font-medium">{$t('notifications.form.enable_title')}</p>
            <p class="text-sm leading-6 text-muted-foreground">
              {$t('notifications.form.enable_description')}
            </p>
          </div>

          <label
            for="smtp_enabled"
            class="flex items-start gap-3 rounded-xl border bg-background px-4 py-3"
          >
            <Checkbox
              id="smtp_enabled"
              bind:checked={draft.enabled}
              disabled={isLoading || isSubmitting}
            />
            <span class="space-y-1">
              <span class="block text-sm font-medium">
                {$t(
                  draft.enabled ? 'notifications.status.enabled' : 'notifications.status.disabled'
                )}
              </span>
              <span class="block text-sm text-muted-foreground">
                {$t('notifications.form.enable_description')}
              </span>
            </span>
          </label>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
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

        <div class="space-y-2 md:col-span-2">
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

      <SMTPModeSelector
        value={draft.security}
        label={$t('notifications.form.transport_title')}
        description={$t('notifications.form.transport_description')}
        options={securityOptions}
        columnsClass="md:grid-cols-3"
        onChange={selectSecurity}
      />

      <SMTPModeSelector
        value={draft.auth_mode}
        label={$t('notifications.form.auth_title')}
        description={$t('notifications.form.auth_description')}
        options={authOptions}
        columnsClass="md:grid-cols-2"
        onChange={selectAuthMode}
      />

      {#if requiresCredentials}
        <div class="grid gap-4 md:grid-cols-2">
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

      {#if draft.security !== 'none'}
        <label
          for="smtp_allow_insecure_tls"
          class="flex items-start gap-3 rounded-xl border bg-background px-4 py-3"
        >
          <Checkbox
            id="smtp_allow_insecure_tls"
            bind:checked={draft.allow_insecure_tls}
            disabled={isLoading || isSubmitting}
          />
          <span class="space-y-1">
            <span class="block text-sm font-medium">
              {$t('notifications.form.allow_insecure_title')}
            </span>
            <span class="block text-sm leading-6 text-muted-foreground">
              {$t('notifications.form.allow_insecure_description')}
            </span>
          </span>
        </label>
      {/if}

      <div class="flex flex-col gap-3 border-t pt-4 md:flex-row md:items-center md:justify-between">
        <p class="text-sm text-muted-foreground">
          {$t('notifications.form.description')}
        </p>

        <div class="flex flex-col gap-2 sm:flex-row">
          <Button
            type="button"
            variant="outline"
            onclick={resetDraft}
            disabled={isLoading || isSubmitting || !isDirty}
          >
            {$t('notifications.form.reset')}
          </Button>
          <Button type="submit" disabled={isLoading || isSubmitting}>
            {$t(isSubmitting ? 'common.saving' : 'notifications.form.save')}
          </Button>
        </div>
      </div>
    </form>
  </Card.Content>
</Card.Root>
