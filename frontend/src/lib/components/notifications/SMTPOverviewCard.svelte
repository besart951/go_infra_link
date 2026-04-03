<script lang="ts">
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import type { SMTPSettings } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  interface Props {
    settings?: SMTPSettings | null;
    isLoading?: boolean;
  }

  let { settings = null, isLoading = false }: Props = $props();

  const t = createTranslator();

  function formatDateTime(value?: string | null): string {
    if (!value) return $t('common.not_available');
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  function securityLabel(security: SMTPSettings['security']): string {
    return $t(`notifications.security.${security}`);
  }

  function authLabel(authMode: SMTPSettings['auth_mode']): string {
    return $t(`notifications.auth.${authMode}`);
  }
</script>

<Card.Root>
  <Card.Header class="gap-3">
    <div class="flex items-start justify-between gap-3">
      <div class="space-y-1">
        <Card.Title>{$t('notifications.overview.title')}</Card.Title>
        <Card.Description>{$t('notifications.overview.description')}</Card.Description>
      </div>
      {#if isLoading}
        <Skeleton class="h-6 w-32 rounded-full" />
      {:else if settings}
        <Badge variant={settings.enabled ? 'default' : 'secondary'}>
          {$t(settings.enabled ? 'notifications.status.enabled' : 'notifications.status.disabled')}
        </Badge>
      {:else}
        <Badge variant="outline">{$t('notifications.status.not_configured')}</Badge>
      {/if}
    </div>
  </Card.Header>

  <Card.Content class="space-y-4">
    {#if isLoading}
      <div class="grid gap-3 sm:grid-cols-2">
        {#each Array(6) as _, index (index)}
          <div class="rounded-xl border bg-background p-4">
            <Skeleton class="h-4 w-28" />
            <Skeleton class="mt-3 h-6 w-40" />
          </div>
        {/each}
      </div>
    {:else if !settings}
      <div class="rounded-xl border border-dashed bg-muted/30 p-4">
        <p class="font-medium">{$t('notifications.overview.empty_title')}</p>
        <p class="mt-1 text-sm leading-6 text-muted-foreground">
          {$t('notifications.overview.empty_description')}
        </p>
      </div>
      <div class="grid gap-3 sm:grid-cols-2">
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.provider')}</p>
          <p class="mt-2 font-medium">{$t('notifications.providers.smtp')}</p>
        </div>
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.transport')}</p>
          <p class="mt-2 font-medium">{$t('notifications.security.starttls')}</p>
        </div>
      </div>
    {:else}
      <div class="flex flex-wrap gap-2">
        <Badge variant="outline">{$t('notifications.providers.smtp')}</Badge>
        <Badge variant="outline">{securityLabel(settings.security)}</Badge>
        <Badge variant="outline">{authLabel(settings.auth_mode)}</Badge>
        <Badge variant="outline">
          {$t(
            settings.has_password
              ? 'notifications.overview.password_set'
              : 'notifications.overview.password_missing'
          )}
        </Badge>
      </div>

      {#if !settings.enabled}
        <Alert.Root>
          <Alert.Description>{$t('notifications.overview.service_disabled')}</Alert.Description>
        </Alert.Root>
      {:else}
        <Alert.Root>
          <Alert.Description>{$t('notifications.overview.service_enabled')}</Alert.Description>
        </Alert.Root>
      {/if}

      <div class="grid gap-3 sm:grid-cols-2">
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.host')}</p>
          <p class="mt-2 font-medium">{settings.host}</p>
        </div>
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.port')}</p>
          <p class="mt-2 font-medium">{settings.port}</p>
        </div>
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.sender')}</p>
          <p class="mt-2 font-medium break-all">{settings.from_email}</p>
          {#if settings.from_name}
            <p class="mt-1 text-sm text-muted-foreground">{settings.from_name}</p>
          {/if}
        </div>
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.reply_to')}</p>
          <p class="mt-2 font-medium break-all">
            {settings.reply_to || $t('common.not_available')}
          </p>
        </div>
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.authentication')}</p>
          <p class="mt-2 font-medium">{authLabel(settings.auth_mode)}</p>
          {#if settings.username}
            <p class="mt-1 text-sm text-muted-foreground">{settings.username}</p>
          {/if}
        </div>
        <div class="rounded-xl border bg-background p-4">
          <p class="text-sm text-muted-foreground">{$t('notifications.overview.updated_at')}</p>
          <p class="mt-2 font-medium">{formatDateTime(settings.updated_at)}</p>
        </div>
      </div>
    {/if}
  </Card.Content>
</Card.Root>
